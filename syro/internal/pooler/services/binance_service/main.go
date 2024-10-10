package binance_service

import (
	"context"
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
	if err := s.setupFuturesAssets(); err != nil {
		s.log().Error(err)
	}

	if err := sched.Register(
		&scheduler.Job{
			Name: "binance-futures-ohlc",
			Freq: "@every 1m",
			// TODO: there is a problem if the custom logger is defined inside the job with
			// the EventID set to the job name. If the function returns error, it won't
			// be associated with it.
			Func: func() error {
				if err := s.runFuturesOhlcScraper(); err != nil {
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

func (s *service) Tmp() {
	if err := s.scrapeFuturesOhlcForId("BTCUSDT", binance.Timeframe15M); err != nil {
		s.log().Error(err)
	}
}

func todoPrinter(v any) { fmt.Println(v) }

func (s *service) getFuturesAssets() ([]market_dto.FuturesAsset, error) {
	coll := s.app.Db().CryptoFuturesAssetColl()

	filter := bson.M{"source": binance.Source, "status": "TRADING"}

	opt := options.Find().
		SetSort(bson.D{{Key: "onboard_date", Value: -1}})

	var docs []market_dto.FuturesAsset
	err := mongodb.GetAllDocumentsWithTypes(coll, filter, opt, &docs)
	return docs, err
}

// If the db does not have any futures assets, scrape the list from the binance api
func (s *service) setupFuturesAssets() error {
	coll := s.app.Db().CryptoFuturesAssetColl()

	filter := bson.M{"source": binance.Source}
	count, err := coll.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}

	if count == 0 {
		s.log().Info("no futures assets found, scraping data")
		return s.scrapeFuturesAssetList()
	}

	s.log().Info(fmt.Sprintf("futures assets already exist in %v collection, skipping setup", coll.Name()))
	return nil
}

func (s *service) runFuturesOhlcScraper(fillgaps ...bool) error {
	assets, err := s.getFuturesAssets()
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
				if err := s.scrapeFuturesOhlcForId(id, binance.Timeframe15M); err != nil {
					s.log().Error(err)
				}
			}

		}(asset.ID)
	}

	wg.Wait()

	return nil
}

func (s *service) scrapeFuturesAssetList() error {
	coll := s.app.Db().CryptoFuturesAssetColl()

	docs, err := s.api.GetAllFutureSymbols()
	if err != nil {
		return err
	}

	log, err := market_dto.UpsertFuturesAssets(docs, coll)
	if err != nil {
		return err
	}

	s.log().Info("upserted binance fututes info", logger.LogFields{"log": log.String()})

	return nil
}

func (s *service) fillGapsForId(id string, tf binance.Timeframe) error {

	coll := s.app.Db().CryptoFuturesOhlcColl()
	filter := bson.M{"id": id, "interval": tf.Milis}
	gaps, err := mongodb.FindGaps(coll, filter)
	if err != nil {
		return err
	}

	if len(gaps) == 0 {
		s.log().Info(fmt.Sprintf("no gaps found for %v", id))
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

			log, err := market_dto.UpsertOhlcRows(docs, coll)
			if err != nil {
				return err
			}

			s.log().Info(fmt.Sprintf("upserted binance fututes ohlc for %v: %v", id, log.String()))
		}
	}

	return nil
}

func (s *service) scrapeFuturesOhlcForId(id string, tf binance.Timeframe) error {
	asset, err := market_dto.GetFuturesAssetByID(s.app.Db().CryptoFuturesAssetColl(), id)
	if err != nil {
		return err
	}

	defaultStart := asset.OnboardDate
	if defaultStart.IsZero() {
		return fmt.Errorf("onboard date is zero for asset %v", id)
	}

	coll := s.app.Db().CryptoFuturesOhlcColl()
	latestStartTime, err := market_dto.GetLatestOhlcStartTime(id, defaultStart, coll, todoPrinter)
	if err != nil {
		return err
	}

	if debugMode {
		// if the latest start time is from the last 3 days, return nil
		breakpoint := time.Now().Add(-3 * 24 * time.Hour)
		if latestStartTime.After(breakpoint) {
			s.log().Info(fmt.Sprintf("latest ohlc for %v is up to date", id))
			return nil
		}
	}

	overlay := tf.CalculateOverlay(10)

	from := latestStartTime.Add(-overlay)
	to := from.Add(tf.GetMaxReqPeriod())

	docs, err := s.api.GetFutureKline(id, from, to, tf)
	if err != nil {
		return err
	}

	log, err := market_dto.UpsertOhlcRows(docs, coll)
	if err != nil {
		return err
	}

	s.log().Info("upserted binance fututes ohlc", logger.LogFields{"id": id, "log": log.String()})

	return nil
}
