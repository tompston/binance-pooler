package binance

import (
	"fmt"
	syro "syro/pkg/app"
	"syro/pkg/models/market_model"
	"testing"
	"time"
)

// go test -run ^TestApi$ syro/internal/pooler/services/binance -v -count=1
func TestApi(t *testing.T) {
	api := NewAPI()
	from := time.Now().Add(-time.Hour * 24).Truncate(time.Hour)
	to := from.Add(time.Hour * 1)

	t.Run("GetFutureKline", func(t *testing.T) {
		dbrows, err := api.GetFutureKline("batusdt", from, to, TIMEFRAME_15M)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(dbrows) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(dbrows))
		}

		for _, row := range dbrows {
			if row.ID != "batusdt" {
				t.Fatalf("expected bat, got %s", row.ID)
			}

			fmt.Printf("GetFutureKline row: %v\n", row)
		}
	})

	t.Run("GetSpotKline - Expected num rows parsed", func(t *testing.T) {
		const past = -time.Hour * 24 * 7

		t1 := time.Now().UTC().Add(past)
		t2 := t1.Add(time.Hour * 1)

		reqPeriod := 60

		api := NewAPI()
		doc, err := api.GetSpotKline("ethusdt", t1, t2, TIMEFRAME_1M)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(doc) != reqPeriod {
			t.Fatalf("expected 60 rows, got %d", len(doc))
		}
	})

}

// GO_CONF_PATH="$(pwd)/conf/config.dev.toml" go test -run ^TestService$ syro/internal/pooler/services//binance -v -count=1
func TestService(t *testing.T) {
	app, cleanup := syro.SetupTestEnvironment(t)
	defer cleanup()

	t.Run("GetFutureKlineTest", func(t *testing.T) {
		api := NewAPI()
		coll := app.Db().TestCollection("crypto_futures_ohlc_service_test")
		from := time.Now().Add(-time.Hour * 24).Truncate(time.Hour)
		to := from.Add(time.Hour * 4)
		id := "batusdt"

		docs, err := api.GetFutureKline(id, from, to, TIMEFRAME_15M)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if doc := market_model.UpsertOhlcRows(docs, coll); doc != nil {
			fmt.Printf("doc.String(): %v\n", doc.String())
		}
	})

	t.Run("scrapeFuturesOhlcForIdTest", func(t *testing.T) {

		s := New(app, 1)

		// need to setup assets first, so that they can be found in the db
		if err := s.setupFuturesAssets(); err != nil {
			s.log().Error(err)
		}

		id := "BTCUSDT"

		if err := s.scrapeFuturesOhlcForId(id, TIMEFRAME_15M); err != nil {
			t.Fatalf(err.Error())
		}
	})
}
