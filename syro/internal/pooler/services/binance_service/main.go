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
	"syro/pkg/lib/scheduler"
	"syro/pkg/providers/binance"

	"syro/pkg/lib/logger"
)

const (
	apiRequestSleep = 500 * time.Millisecond
	debugMode       = true
)

type service struct {
	app                 *app.App
	api                 binance.API
	maxParalellRequests int
}

func New(app *app.App, maxParalellRequests int) *service {
	return &service{app, binance.NewAPI(), maxParalellRequests}
}

func (s *service) log() logger.Logger {
	return s.app.Logger().SetEvent("binance")
}

func (s *service) AddJobs(sched *scheduler.Scheduler) error {
	// if err := s.setupFuturesAssets(); err != nil {
	// 	s.log().Error(err)
	// }

	if err := s.setupSpotAssets(); err != nil {
		s.log().Error(err)
	}

	if err := sched.Register(
		&scheduler.Job{
			Name: "binance-spot-ohlc",
			Freq: "@every 1m",
			Func: func() error {
				if err := s.runOhlcScraper(); err != nil {
					s.log().Error(err)
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

// func (s *service) Tmp() {
// 	if err := s.scrapeOhlcForID("BTCUSDT", binance.Timeframe15M); err != nil {
// 		s.log().Error(err)
// 	}
// }

func (s *service) runOhlcScraper(fillgaps ...bool) error {

	assets, err := market_dto.NewMongoInterface().GetSpotAssets(
		s.app.Db().CryptoSpotAssetColl(),
		bson.M{"source": binance.Source, "status": "TRADING"},
		// options.Find().SetSort(bson.D{{Key: "onboard_date", Value: -1}}),
		options.Find().SetLimit(50),
	)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, s.maxParalellRequests)
	var wg sync.WaitGroup

	for _, asset := range assets {
		sem <- struct{}{}
		wg.Add(1)

		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			time.Sleep(apiRequestSleep)

			if len(fillgaps) == 1 && fillgaps[0] {
				if err := s.fillGapsForId(id, binance.Timeframe15M); err != nil {
					s.log().Error(err)
				}

			} else {
				if err := s.scrapeOhlcForID(id, binance.Timeframe15M); err != nil {
					s.log().Error(err)
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
		s.log().Info("no gaps found for futures ohlc", logger.Fields{"id": id})
		return nil
	}

	for _, gap := range gaps {

		for _, g := range gap {
			from := g.StartOfGap
			to := g.EndOfGap
			docs, err := s.api.GetFutureKline(id, from, to, tf)
			if err != nil {
				return err
			}

			log, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
			if err != nil {
				return err
			}

			s.log().Info(fmt.Sprintf("upserted binance fututes ohlc for %v: %v", id, log.String()))
		}
	}

	return nil
}

func (s *service) scrapeOhlcForID(id string, tf binance.Timeframe) error {
	// asset, err := market_dto.NewMongoInterface().GetSpotAssetByID(s.app.Db().CryptoSpotAssetColl(), id)
	// if err != nil {
	// 	return err
	// }

	// defaultStart := asset.OnboardDate
	// if defaultStart.IsZero() {
	// 	return fmt.Errorf("onboard date is zero for asset %v", id)
	// }

	defaultStart := time.Now().AddDate(-4, 0, 0)

	coll := s.app.Db().CryptoSpotOhlcColl()
	latestStartTime, err := market_dto.NewMongoInterface().GetLatestOhlcStartTime(id, defaultStart, coll, nil)
	if err != nil {
		return err
	}

	if debugMode {
		// if the latest start time is from the last 3 days, return nil
		breakpoint := time.Now().Add(-3 * 24 * time.Hour)
		if latestStartTime.After(breakpoint) {
			s.log().Info("latest ohlc for is up to date", logger.Fields{"id": id})
			return nil
		}
	}

	overlay := tf.CalculateOverlay(10)

	from := latestStartTime.Add(-overlay)
	to := from.Add(tf.GetMaxReqPeriod())

	docs, err := s.api.GetSpotKline(id, from, to, tf)
	if err != nil {
		return err
	}

	log, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
	if err != nil {
		return fmt.Errorf("%v:%v failed to upsert ohlc rows: %v", id, tf.UrlParam, err)
	}

	s.log().Info("upserted binance fututes ohlc", logger.Fields{"id": id, "log": log.String()})

	return nil
}
