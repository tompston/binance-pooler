package binance_service

import (
	"binance-pooler/pkg/app"
	"binance-pooler/pkg/dto/market_dto"
	"binance-pooler/pkg/providers/binance"
	"fmt"
	"testing"
	"time"
)

func TestApi(t *testing.T) {
	api := binance.NewAPI()
	from := time.Now().Add(-time.Hour * 24).Truncate(time.Hour)
	to := from.Add(time.Hour * 1)

	t.Run("GetFutureKline", func(t *testing.T) {
		dbrows, err := api.GetFutureKline("batusdt", from, to, binance.Timeframe15M)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(dbrows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(dbrows))
		}

		for _, row := range dbrows {
			if row.Symbol != "batusdt" {
				t.Fatalf("expected bat, got %s", row.Symbol)
			}

			fmt.Printf("GetFutureKline row: %v\n", row)
		}
	})

	t.Run("GetSpotKline - Expected num rows parsed", func(t *testing.T) {
		const past = -time.Hour * 24 * 7

		t1 := time.Now().UTC().Add(past)
		t2 := t1.Add(time.Hour * 1)

		reqPeriod := 60

		api := binance.NewAPI()
		doc, err := api.GetSpotKline("ethusdt", t1, t2, binance.Timeframe1M)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(doc) != reqPeriod {
			t.Fatalf("expected 60 rows, got %d", len(doc))
		}
	})
}

func TestService(t *testing.T) {
	app, cleanup := app.SetupTestEnvironment(t)
	defer cleanup()

	mongoInterface := market_dto.NewMongoInterface()

	t.Run("GetFutureKlineTest", func(t *testing.T) {
		api := binance.NewAPI()
		coll := app.Db().TestCollection("crypto_futures_ohlc_service_test")
		from := time.Now().Add(-time.Hour * 24).Truncate(time.Hour)
		to := from.Add(time.Hour * 4)
		symbol := "batusdt"

		docs, err := api.GetFutureKline(symbol, from, to, binance.Timeframe15M)
		if err != nil {
			t.Fatalf(err.Error())
		}

		log, err := mongoInterface.UpsertOhlcRows(docs, coll)
		if err != nil {
			t.Fatalf(err.Error())
		}

		fmt.Printf("log.String(): %v\n", log.String())
	})

	t.Run("scrapeOhlcForSymbolTest", func(t *testing.T) {

		s := New(app, 1)

		// need to setup assets first, so that they can be found in the db
		if err := s.setupFuturesAssets(); err != nil {
			s.log().Error(err.Error())
		}

		if err := s.scrapeOhlcForSymbol("BTCUSDT", binance.Timeframe15M); err != nil {
			t.Fatalf(err.Error())
		}
	})
}
