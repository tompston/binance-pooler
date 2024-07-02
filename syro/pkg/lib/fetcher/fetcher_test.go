package fetcher

import (
	"testing"
)

func TestFetch(t *testing.T) {
	if _, err := Fetch("GET", "https://httpbin.org/get", nil); err != nil {
		t.Errorf("error fetching %v", err)
	}
}
