package binance_service

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"syro/pkg/app"
	"syro/pkg/dto/market_dto"
	"syro/pkg/lib/mongodb"
	"syro/pkg/providers/binance"
	"syro/pkg/sy"
	"syro/pkg/sy/timeset"
)

const (
	apiRequestSleep = 500 * time.Millisecond
)

type service struct {
	app                 *app.App
	api                 binance.API
	maxParalellRequests int
	debugMode           bool
}

func New(app *app.App, maxParalellRequests int) *service {
	return &service{
		maxParalellRequests: maxParalellRequests,
		api:                 binance.NewAPI(),
		debugMode:           false,
		app:                 app,
	}
}

// Set debug mode to true
func (s *service) WithDebugMode() *service {
	s.debugMode = true
	return s
}

func (s *service) log() sy.Logger {
	return s.app.Logger().SetEvent("binance")
}

func (s *service) AddJobs(sched *sy.CronScheduler) error {
	if err := s.setupSpotAssets(); err != nil {
		s.log().Fatal(err.Error())
	}

	if err := sched.Register(
		&sy.Job{
			Name: "binance-spot-ohlc",
			Freq: "@every 15s",
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
	// if err := s.scrapeOhlcForID("BTCUSDT", binance.Timeframe15M); err != nil {
	// s.log().Error(err)
	// }

	if err := s.runOhlcScraper(fill); err != nil {
		s.log().Error(err.Error())
	}
}

func (s *service) getTopPairs() ([]market_dto.SpotAsset, error) {
	return market_dto.NewMongoInterface().GetSpotAssets(
		s.app.Db().CryptoSpotAssetColl(),
		bson.M{"source": binance.Source, "id": bson.M{"$in": binance.TopPairs}}, nil,
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

	s.log().Debug(" running ohlc scraper", sy.LogFields{"num_assets": len(assets)})

	for _, asset := range assets {
		sem <- struct{}{}
		wg.Add(1)

		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			time.Sleep(apiRequestSleep)

			if fillgaps {
				if err := s.fillGapsForId(id, binance.Timeframe15M); err != nil {
					s.log().Error(err.Error())
				}

			} else {
				if err := s.scrapeOhlcForID(id, binance.Timeframe15M); err != nil {
					s.log().Error(err.Error())
				}
			}

		}(asset.ID)
	}

	wg.Wait()

	return nil
}

func (s *service) fillGapsForId(id string, tf binance.Timeframe) error {

	coll := s.app.Db().CryptoSpotOhlcColl()
	filter := bson.M{"id": id, "interval": tf.Milis}
	gaps, err := mongodb.FindGaps(coll, filter)
	if err != nil {
		return err
	}

	if len(gaps) == 0 {
		s.log().Debug("no gaps found for futures ohlc", sy.LogFields{"id": id})
		return nil
	}

	for interval, gap := range gaps {

		s.log().Debug("found gaps for interval", sy.LogFields{"id": id, "interval": interval, "gaps": len(gap)})

		for _, g := range gap {
			s.log().Debug("filling gap", sy.LogFields{"id": id, "gap": g.String()})

			// The gaps might exceed the 1k limit of the api, that's why we chunk the time range
			// into smaller pieces and request them one by one.
			gapChunks := timeset.ChunkTimeRange(g.StartOfGap, g.EndOfGap, timeset.MilisToDuration(interval), 500, 10)

			s.log().Debug("period chunks", sy.LogFields{"chunks": len(gapChunks), "id": id})

			for chunkIdx, chunk := range gapChunks {
				s.log().Debug(fmt.Sprintf("requesting chunk [%v / %v] for %v from %v -> %v", chunkIdx, len(gapChunks), id, chunk.From, chunk.To))

				docs, err := s.api.GetFutureKline(id, chunk.From, chunk.To, tf)
				if err != nil {
					return fmt.Errorf("%v:%v [%v -> %v] failed to get ohlc rows: %v", id, tf.UrlParam, chunk.From, chunk.To, err)
				}

				upsertLog, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
				if err != nil {
					return err
				}

				s.log().Info("upserted binance fututes ohlc", sy.LogFields{"id": id, "log": upsertLog.String()})
			}
		}
	}

	return nil
}

func (s *service) scrapeOhlcForID(id string, tf binance.Timeframe) error {
	defaultStart := time.Now().AddDate(-4, 0, 0)

	coll := s.app.Db().CryptoSpotOhlcColl()
	latestTime, err := market_dto.NewMongoInterface().GetLatestOhlcStartTime(id, defaultStart, coll, nil)
	if err != nil {
		return err
	}

	if s.debugMode {
		// if the latest start time is from the last 3 days, return nil
		breakpoint := time.Now().AddDate(0, 0, -1)
		if latestTime.After(breakpoint) {
			s.log().Info("latest ohlc for is up to date", sy.LogFields{"id": id})
			return nil
		}
	}

	overlay := tf.CalculateOverlay(10)

	from := latestTime.Add(-overlay)
	to := from.Add(tf.GetMaxReqPeriod())

	docs, err := s.api.GetSpotKline(id, from, to, tf)
	if err != nil {
		return err
	}

	upsertLog, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
	if err != nil {
		return fmt.Errorf("%v:%v failed to upsert ohlc rows: %v", id, tf.UrlParam, err)
	}

	s.log().Info("upserted binance fututes ohlc", sy.LogFields{"id": id, "upsertLog": upsertLog.String()})

	return nil
}
