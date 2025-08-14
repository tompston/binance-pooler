package binance_service

import (
	"binance-pooler/pkg/app"
	"binance-pooler/pkg/providers/binance"
	"fmt"
	"testing"
	"time"
)

func TestApi(t *testing.T) {

	t.Run("GetSpotKline - Expected num rows parsed", func(t *testing.T) {
		const past = -time.Hour * 24 * 7

		t1 := time.Now().UTC().Add(past)
		t2 := t1.Add(time.Hour * 1)

		reqPeriod := 60

		api := binance.New()
		doc, err := api.GetSpotKline("ethusdt", t1, t2, binance.Timeframe1M)
		if err != nil {
			t.Fatal(err)
		}

		if len(doc) != reqPeriod {
			t.Fatalf("expected 60 rows, got %d", len(doc))
		}
	})
}

func TestService(t *testing.T) {
	app, cleanup := app.SetupTestEnvironment(t)
	defer cleanup()

	t.Run("get-spot-kline-flow", func(t *testing.T) {
		api := binance.New()
		coll := app.Db().TestCollection("crypto_spot_ohlc_service_test")
		from := time.Now().Add(-time.Hour * 24).Truncate(time.Hour)
		to := from.Add(time.Hour * 4)
		symbol := "batusdt"

		docs, err := api.GetSpotKline(symbol, from, to, binance.Timeframe15M)
		if err != nil {
			t.Fatal(err)
		}

		log, err := marketdb.UpsertOhlcRows(docs, coll)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("log.String(): %v\n", log.String())
	})

	t.Run("scrapeOhlcForSymbolTest", func(t *testing.T) {

		s := New(app, 1, []binance.Timeframe{})

		// need to setup assets first, so that they can be found in the db
		if err := s.setupSpotAssets(); err != nil {
			s.log().Error(err.Error())
		}

		if err := s.scrapeOhlcForSymbol("BTCUSDT", binance.Timeframe15M); err != nil {
			t.Fatal(err)
		}
	})
}
