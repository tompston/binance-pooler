package binance_service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"binance-pooler/pkg/app"
	"binance-pooler/pkg/dto/market_dto"
	"binance-pooler/pkg/lib/mongodb"
	"binance-pooler/pkg/providers/binance"

	"binance-pooler/pkg/lib/timeset"

	"github.com/tompston/syro"
)

type service struct {
	app                  *app.App
	api                  binance.API
	maxParallelRequests  int
	timeframes           []binance.Timeframe
	requestSleepDuration time.Duration
	debug                bool
}

func New(app *app.App, maxParallelRequests int, timeframes []binance.Timeframe) *service {
	return &service{
		maxParallelRequests: maxParallelRequests,
		api:                 binance.New(),
		timeframes:          timeframes,
		debug:               false,
		app:                 app,
	}
}

func (s *service) WithDebug() *service {
	s.debug = true
	return s
}

func (s *service) WithSleepDuration(d time.Duration) *service {
	s.requestSleepDuration = d
	return s
}

func (s *service) log() syro.Logger {
	return s.app.Logger().WithEvent("binance")
}

func (s *service) AddJobs(sched *syro.CronScheduler) error {
	if err := s.setupSpotAssets(); err != nil {
		return err
	}

	if err := s.setupFuturesAssets(); err != nil {
		return err
	}

	if err := sched.Register(
		&syro.Job{
			Name:     "binance-spot-ohlc",
			Schedule: "@every 30s",
			Func: func() error {
				assetsColl := s.app.Db().CryptoSpotAssetColl()
				historyColl := s.app.Db().CryptoSpotOhlcColl()
				getFunc := s.api.GetSpotKline

				filter := bson.M{"source": binance.Source, "symbol": bson.M{"$in": binance.TopPairs}}

				assets, err := market_dto.GetAssets(assetsColl, filter, nil)
				if err != nil {
					s.log().Error(err.Error())
					return err
				}

				if err := s.runOhlcScraper(assets, historyColl, getFunc, false); err != nil {
					s.log().Error(err.Error())
					return err
				}

				return nil
			},
		},
	); err != nil {
		return err
	}

	if err := sched.Register(
		&syro.Job{
			Name:     "binance-futures-ohlc",
			Schedule: "@every 30s",
			Func: func() error {
				assetsColl := s.app.Db().CryptoFuturesAssetColl()
				historyColl := s.app.Db().CryptoFuturesOhlcColl()
				getFunc := s.api.GetFutureKline

				filter := bson.M{"source": binance.Source, "symbol": bson.M{"$in": binance.TopPairs}}

				assets, err := market_dto.GetAssets(assetsColl, filter, nil)
				if err != nil {
					s.log().Error(err.Error())
					return err
				}

				if err := s.runOhlcScraper(assets, historyColl, getFunc, false); err != nil {
					s.log().Error(err.Error())
					return err
				}

				return nil
			},
		},
	); err != nil {
		return err
	}

	return nil
}

func (s *service) setupSpotAssets() error {
	assetsColl := s.app.Db().CryptoSpotAssetColl()
	getFunc := s.api.GetAllSpotAssets
	return initializeAssets(s, assetsColl, getFunc)
}

func (s *service) setupFuturesAssets() error {
	assetsColl := s.app.Db().CryptoFuturesAssetColl()
	getFunc := s.api.GetAllFutureSymbols
	return initializeAssets(s, assetsColl, getFunc)
}

func initializeAssets[T any](s *service, assetsColl *mongo.Collection, getAssets binance.GetAssetsFunc[T]) error {
	filter := bson.M{"source": binance.Source}
	count, err := assetsColl.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}

	if count == 0 {
		s.log().Info("no assets found, scraping data", syro.LogFields{"collection": assetsColl.Name()})
		docs, err := getAssets()
		if err != nil {
			return err
		}

		upsertLog, err := market_dto.UpsertAssets(docs, assetsColl)
		if err != nil {
			return err
		}

		s.log().Info("upserted binance asset info", syro.LogFields{"upsertLog": upsertLog})
		return nil
	}

	s.log().Info(fmt.Sprintf("assets already exist in %v collection, skipping setup", assetsColl.Name()))
	return nil
}

func (s *service) runOhlcScraper(assets []market_dto.AssetBase, coll *mongo.Collection, getHistoryFunc binance.GetHistoryFunc, fillgaps bool) error {

	sem := make(chan struct{}, s.maxParallelRequests)
	var wg sync.WaitGroup

	s.log().Debug("running ohlc scraper", syro.LogFields{"num_assets": len(assets), "coll": coll.Name()})

	for _, asset := range assets {
		sem <- struct{}{}
		wg.Add(1)

		go func(symbol string) {
			defer wg.Done()
			defer func() { <-sem }()

			for _, tf := range s.timeframes {
				time.Sleep(s.requestSleepDuration)
				if fillgaps {
					if err := s.fillGapsForSymbol(coll, getHistoryFunc, symbol, tf); err != nil {
						s.log().Error(err.Error())
					}

				} else {
					if err := s.scrapeOhlcForSymbol(coll, getHistoryFunc, symbol, tf); err != nil {
						s.log().Error(err.Error())
					}
				}
			}

		}(asset.Symbol)
	}

	wg.Wait()

	return nil
}

func (s *service) fillGapsForSymbol(historyColl *mongo.Collection, getHistoryFunc binance.GetHistoryFunc, symbol string, tf binance.Timeframe) error {
	filter := bson.M{"symbol": symbol, "interval": tf.Milis}
	gaps, err := mongodb.FindGaps(historyColl, filter)
	if err != nil {
		return err
	}

	if len(gaps) == 0 {
		s.log().Debug("no gaps found for ohlc", syro.LogFields{"symbol": symbol, "interval": tf.Milis})
		return nil
	}

	for interval, gap := range gaps {

		s.log().Debug("found gaps", syro.LogFields{"symbol": symbol, "interval": interval, "num_gaps": len(gap)})

		for _, g := range gap {
			s.log().Debug("filling gap", syro.LogFields{"symbol": symbol, "gap": g.String()})

			// The gaps might exceed the 1k limit of the api, that's why we chunk the time range
			// into smaller pieces and request them one by one.
			gapChunks, err := timeset.ChunkTimeRange(g.StartOfGap, g.EndOfGap, timeset.MilisToDuration(interval), 500, 10)
			if err != nil {
				return err
			}

			s.log().Debug("period chunks", syro.LogFields{"chunks": len(gapChunks), "symbol": symbol})

			for chunkIdx, chunk := range gapChunks {

				s.log().Debug("requesting chunk", syro.LogFields{
					"chunk_idx":  chunkIdx,
					"num_chunks": len(gapChunks),
					"symbol":     symbol,
					"from":       chunk.From,
					"to":         chunk.To,
					"interval":   tf.Milis,
				})

				docs, err := getHistoryFunc(symbol, chunk.From, chunk.To, tf)
				if err != nil {
					return fmt.Errorf("%v:%v [%v -> %v] failed to get ohlc rows: %v", symbol, tf.UrlParam, chunk.From, chunk.To, err)
				}

				upsertLog, err := market_dto.UpsertOhlcRows(docs, historyColl)
				if err != nil {
					return err
				}

				s.log().Info("upserted ohlc", syro.LogFields{"symbol": symbol, "log": upsertLog})
			}
		}
	}

	return nil
}

func (s *service) scrapeOhlcForSymbol(historyColl *mongo.Collection, getHistoryFunc binance.GetHistoryFunc, symbol string, tf binance.Timeframe) error {
	defaultStart := time.Now().AddDate(-6, 0, 0)

	filter := bson.M{
		"interval": tf.Milis,
		"symbol":   symbol,
	}

	latestTime, err := mongodb.FindLatestStartTime(defaultStart, historyColl, filter)
	if err != nil {
		return err
	}

	if s.debug {
		// if the latest start time is from the last x days, return nil
		breakpoint := time.Now().AddDate(0, 0, -1)
		if latestTime.After(breakpoint) {
			s.log().Info("latest ohlc is up to date", syro.LogFields{"symbol": symbol, "interval": tf.Milis})
			return nil
		}
	}

	overlay := tf.CalculateOverlay(20)
	maxPeriod := tf.GetMaxReqPeriod()

	// a dumb but simple way to handle the following problem: we don't know from which date
	// the data is available, so we iterate from the default start time, untill time series
	// data is found. This block executes only on the first run.
	if latestTime.Equal(defaultStart) {

		current := latestTime
		stop := time.Now()

		for current.Before(stop) {
			from := current.Add(-overlay)
			to := from.Add(maxPeriod)

			time.Sleep(s.requestSleepDuration)

			meta := syro.LogFields{"symbol": symbol, "resolution": tf.Milis / 60000, "from": from.UTC(), "to": to.UTC()}

			s.log().Debug("init request ohlc", meta)

			docs, err := getHistoryFunc(symbol, from, to, tf)
			if err != nil {
				return fmt.Errorf("%v:%v failed to get ohlc rows: %v", symbol, tf.UrlParam, err)
			}

			if len(docs) != 0 {
				upsertLog, err := market_dto.UpsertOhlcRows(docs, historyColl)
				if err != nil {
					return fmt.Errorf("%v:%v failed to upsert ohlc rows: %v", symbol, tf.UrlParam, err)
				}

				s.log().Info("upserted binance ohlc",
					syro.LogFields{
						"symbol":     symbol,
						"resolution": tf.Milis / 60000,
						"upsertLog":  upsertLog.String(),
					})
				break

			} else {
				s.log().Debug("no ohlc data found", meta)
				current = current.Add(maxPeriod)
			}
		}

	} else {
		from := latestTime.Add(-overlay)
		to := from.Add(maxPeriod)

		docs, err := getHistoryFunc(symbol, from, to, tf)
		if err != nil {
			return err
		}

		upsertLog, err := market_dto.UpsertOhlcRows(docs, historyColl)
		if err != nil {
			return fmt.Errorf("%v:%v failed to upsert ohlc rows: %v", symbol, tf.UrlParam, err)
		}

		s.log().Info("upserted binance ohlc",
			syro.LogFields{
				"symbol":     symbol,
				"resolution": tf.Milis / 60000,
				"upsertLog":  upsertLog.String(),
				"coll":       historyColl.Name(),
			})
	}

	return nil
}
