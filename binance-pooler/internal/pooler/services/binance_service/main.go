package binance_service

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"binance-pooler/pkg/app"
	"binance-pooler/pkg/dto/market_dto"
	"binance-pooler/pkg/lib/mongodb"
	"binance-pooler/pkg/providers/binance"
	"binance-pooler/pkg/syro"
	"binance-pooler/pkg/syro/timeset"
)

type service struct {
	app                  *app.App
	api                  binance.API
	maxParalellRequests  int
	timeframes           []binance.Timeframe
	requestSleepDuration time.Duration
	debugMode            bool
}

func New(app *app.App, maxParalellRequests int, timeframes []binance.Timeframe) *service {
	return &service{
		maxParalellRequests: maxParalellRequests,
		api:                 binance.NewAPI(),
		timeframes:          timeframes,
		debugMode:           false,
		app:                 app,
	}
}

// Set debug mode to true
func (s *service) WithDebugMode() *service {
	s.debugMode = true
	return s
}

func (s *service) WithSleepDuration(d time.Duration) *service {
	s.requestSleepDuration = d
	return s
}

func (s *service) log() syro.Logger {
	return s.app.Logger().SetEvent("binance")
}

func (s *service) AddJobs(sched *syro.CronScheduler) error {
	if err := s.setupSpotAssets(); err != nil {
		s.log().Fatal(err.Error())
	}

	if err := sched.Register(
		&syro.Job{
			Name: "binance-spot-ohlc",
			Freq: "@every 30s",
			Func: func() error {
				if err := s.runOhlcScraper(false); err != nil {
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

func (s *service) Tmp(fill bool) {
	if err := s.runOhlcScraper(fill); err != nil {
		s.log().Error(err.Error())
	}
}

func (s *service) getTopPairs() ([]market_dto.SpotAsset, error) {
	return market_dto.NewMongoInterface().GetSpotAssets(
		s.app.Db().CryptoSpotAssetColl(),
		bson.M{"source": binance.Source, "symbol": bson.M{"$in": binance.TopPairs}}, nil,
	)
}

func (s *service) getAllTradingPairs(limit int64) ([]market_dto.SpotAsset, error) {
	return market_dto.NewMongoInterface().GetSpotAssets(
		s.app.Db().CryptoSpotAssetColl(),
		bson.M{"source": binance.Source, "status": "TRADING"},
		options.Find().SetLimit(limit), // options.Find().SetSort(bson.D{{Key: "onboard_date", Value: -1}}),
	)
}

func (s *service) runOhlcScraper(fillgaps bool) error {
	assets, err := s.getTopPairs()
	if err != nil {
		return err
	}

	sem := make(chan struct{}, s.maxParalellRequests)
	var wg sync.WaitGroup

	s.log().Debug("* running ohlc scraper", syro.LogFields{"num_assets": len(assets)})

	for _, asset := range assets {
		sem <- struct{}{}
		wg.Add(1)

		go func(symbol string) {
			defer wg.Done()
			defer func() { <-sem }()

			for _, tf := range s.timeframes {
				time.Sleep(s.requestSleepDuration)
				if fillgaps {
					if err := s.fillGapsForSymbol(symbol, tf); err != nil {
						s.log().Error(err.Error())
					}

				} else {
					if err := s.scrapeOhlcForSymbol(symbol, tf); err != nil {
						s.log().Error(err.Error())
					}
				}
			}

		}(asset.Symbol)
	}

	wg.Wait()

	return nil
}

func (s *service) fillGapsForSymbol(symbol string, tf binance.Timeframe) error {

	coll := s.app.Db().CryptoSpotOhlcColl()
	filter := bson.M{"symbol": symbol, "interval": tf.Milis}
	gaps, err := mongodb.FindGaps(coll, filter)
	if err != nil {
		return err
	}

	if len(gaps) == 0 {
		s.log().Debug("no gaps found for futures ohlc", syro.LogFields{"symbol": symbol, "interval": tf.Milis})
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

				s.log().Debug(fmt.Sprintf("requesting chunk %v/%v for %v from %v -> %v", chunkIdx, len(gapChunks), symbol, chunk.From, chunk.To))

				docs, err := s.api.GetFutureKline(symbol, chunk.From, chunk.To, tf)
				if err != nil {
					return fmt.Errorf("%v:%v [%v -> %v] failed to get ohlc rows: %v", symbol, tf.UrlParam, chunk.From, chunk.To, err)
				}

				upsertLog, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
				if err != nil {
					return err
				}

				s.log().Info("upserted binance fututes ohlc", syro.LogFields{"symbol": symbol, "log": upsertLog.String()})
			}
		}
	}

	return nil
}

func (s *service) scrapeOhlcForSymbol(symbol string, tf binance.Timeframe) error {

	defaultStart := time.Now().AddDate(-5, 0, 0)
	coll := s.app.Db().CryptoSpotOhlcColl()

	filter := bson.M{
		"interval": tf.Milis,
		"symbol":   symbol,
	}

	latestTime, err := mongodb.GetLatestStartTime(defaultStart, coll, filter, false)
	if err != nil {
		return err
	}

	if s.debugMode {
		// if the latest start time is from the last 3 days, return nil
		breakpoint := time.Now().AddDate(0, 0, -1)
		if latestTime.After(breakpoint) {
			s.log().Info("latest ohlc is up to date", syro.LogFields{"symbol": symbol, "interval": tf.Milis})
			return nil
		}
	}

	overlay := tf.CalculateOverlay(20)

	from := latestTime.Add(-overlay)
	to := from.Add(tf.GetMaxReqPeriod())

	docs, err := s.api.GetSpotKline(symbol, from, to, tf)
	if err != nil {
		return err
	}

	upsertLog, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
	if err != nil {
		return fmt.Errorf("%v:%v failed to upsert ohlc rows: %v", symbol, tf.UrlParam, err)
	}

	s.log().Info("upserted binance fututes ohlc",
		syro.LogFields{"symbol": symbol, "resolution": tf.Milis / 60000, "upsertLog": upsertLog.String()})

	return nil
}
