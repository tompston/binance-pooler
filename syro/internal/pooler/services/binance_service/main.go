package binance_service

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"syro/pkg/app"
	"syro/pkg/dto/market_dto"
	"syro/pkg/lib/logbook"
	"syro/pkg/lib/mongodb"
	"syro/pkg/lib/scheduler"
	"syro/pkg/lib/timeset"
	"syro/pkg/providers/binance"
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

func (s *service) log() logbook.Logger {
	return s.app.Logger().SetEvent("binance")
}

func (s *service) AddJobs(sched *scheduler.Scheduler) error {
	// if err := s.setupFuturesAssets(); err != nil {
	// 	s.log().Error(err)
	// }

	if err := s.setupSpotAssets(); err != nil {
		s.log().Error(err.Error())
	}

	if err := sched.Register(
		&scheduler.Job{
			Name: "binance-spot-ohlc",
			Freq: "@every 1m",
			Func: func() error {
				if err := s.runOhlcScraper(); err != nil {
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

func (s *service) Tmp() {
	// if err := s.scrapeOhlcForID("BTCUSDT", binance.Timeframe15M); err != nil {
	// s.log().Error(err)
	// }

	if err := s.runOhlcScraper(true); err != nil {
		s.log().Error(err.Error())
	}
}

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

	s.log().Debug("* running ohlc scraper", logbook.Fields{"count": len(assets)})

	for _, asset := range assets {
		sem <- struct{}{}
		wg.Add(1)

		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			time.Sleep(apiRequestSleep)

			if len(fillgaps) == 1 && fillgaps[0] {
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
		s.log().Info("no gaps found for futures ohlc", logbook.Fields{"id": id})
		return nil
	}

	for interval, gap := range gaps {

		_ = interval

		s.log().Debug(
			"found gaps for interval",
			logbook.Fields{"id": id, "interval": interval, "gaps": len(gap)},
		)

		for _, g := range gap {

			s.log().Info("filling gap", logbook.Fields{"id": id, "gap": g.String()})

			gapChunks := timeset.ChunkTimeRange(g.StartOfGap, g.EndOfGap, timeset.MilisToDuration(interval), 500, 10)

			s.log().Debug("period chunks", logbook.Fields{"chunks": len(gapChunks)})

			for chunkIdx, chunk := range gapChunks {

				s.log().Debug(" - requesting chunk", logbook.Fields{"chunk": chunkIdx, "from": chunk.From, "to": chunk.To, "id": id})

				docs, err := s.api.GetFutureKline(id, chunk.From, chunk.To, tf)
				if err != nil {
					return err
				}

				upsertLog, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
				if err != nil {
					return err
				}

				s.log().Info("upserted binance fututes ohlc",
					logbook.Fields{"id": id, "log": upsertLog.String()})
			}

			docs, err := s.api.GetFutureKline(id, g.EndOfGap, g.StartOfGap, tf)
			if err != nil {
				return err
			}

			upsertLog, err := market_dto.NewMongoInterface().UpsertOhlcRows(docs, coll)
			if err != nil {
				return err
			}

			s.log().Info("upserted binance fututes ohlc",
				logbook.Fields{"id": id, "log": upsertLog.String()})

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
			s.log().Info("latest ohlc for is up to date", logbook.Fields{"id": id})
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

	s.log().Info("upserted binance fututes ohlc", logbook.Fields{"id": id, "log": log.String()})

	return nil
}
