package binance

import (
	"testing"
	"time"
)

func TestApi(t *testing.T) {

	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Hour * 4)
	timerfame := Timeframe15M
	symbol := "ethusdt"

	t.Run("GetSpotKline", func(t *testing.T) {
		docs, err := New().GetSpotKline(symbol, t1, t2, timerfame)
		if err != nil {
			t.Fatal(err)
		}

		if len(docs) == 0 {
			t.Fatalf("expected 5 rows, got %d", len(docs))
		}
	})

	/*
		t.Run("GetFutureKline", func(t *testing.T) {
			docs, err := New().GetFutureKline(symbol, t1, t2, timerfame)
			if err != nil {
				t.Fatal(err)
			}

			if len(docs) == 0 {
				t.Fatalf("expected 5 rows, got %d", len(docs))
			}
		})
	*/
}
