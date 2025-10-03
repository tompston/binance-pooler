package binance

import "time"

const Source = "binance"

type API struct{}

func New() API { return API{} }

var TopPairs = []string{
	"BTCUSDT",
	"ETHUSDT",
	"SOLUSDT",
	"BNBUSDT",
	"AVAXUSDT",
	"LINKUSDT",
	"ADAUSDT",
	// "DOGEUSDT",
	// "XRPTUSDT",
	// "TRXUSDT",
	// "LTCUSDT",
}

type Timeframe struct {
	UrlParam string
	Milis    int64
}

const minInMillis = 60 * 1000

var (
	Timeframe1M  = Timeframe{"1m", 1 * minInMillis}
	Timeframe5M  = Timeframe{"5m", 5 * minInMillis}
	Timeframe15M = Timeframe{"15m", 15 * minInMillis}
	Timeframe30M = Timeframe{"30m", 30 * minInMillis}
	Timeframe1H  = Timeframe{"1h", 60 * minInMillis}
)

// GetMaxReqPeriod returns the maximum period that can be requested from the
// binance api, based on the requested resolution of the data. The api has a
// limit of 1000 data points per request.
func (tf Timeframe) GetMaxReqPeriod() time.Duration {
	return time.Duration(0.97*float64(tf.Milis)*1000) * time.Millisecond
}

// CalculateOverlay returns the time duration that should be added to the
// start time of the request in order to avoid gaps in the data.
func (tf Timeframe) CalculateOverlay(numEntries int64) time.Duration {
	return time.Duration(numEntries*tf.Milis) * time.Millisecond
}
