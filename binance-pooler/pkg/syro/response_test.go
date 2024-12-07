package syro

import (
	"testing"
	"time"
)

func TestResponse(t *testing.T) {

	t.Run("parseLogsQuery", func(t *testing.T) {
		url := "http://localhost:8080/logs?from=2021-01-01T00:00:00Z&to=2021-01-02T00:00:00Z&limit=10&skip=5"

		filter, err := parseLogsQuery(url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if filter.TimeseriesFilter.From != time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC) {
			t.Fatalf("unexpected 'from' time: %v", filter.TimeseriesFilter.From)
		}

		if filter.TimeseriesFilter.To != time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC) {
			t.Fatalf("unexpected 'to' time: %v", filter.TimeseriesFilter.To)
		}

		if filter.TimeseriesFilter.Limit != 10 {
			t.Fatalf("unexpected 'limit' value: %v", filter.TimeseriesFilter.Limit)
		}
	})
}
