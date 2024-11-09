package binance

import (
	"fmt"
	"strconv"
	"strings"
	"syro/pkg/dto/market_dto"
	"syro/pkg/lib/encoder"
	"syro/pkg/lib/fetcher"
	"time"
)

const Source = "binance"

type API struct{}

func NewAPI() API { return API{} }

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

type PeriodChunk struct {
	From time.Time
	To   time.Time
}

// CalculateOverlay returns the time duration that should be added to the
// start time of the request in order to avoid gaps in the data.
func (tf Timeframe) CalculateOverlay(numEntries int64) time.Duration {
	return time.Duration(numEntries*tf.Milis) * time.Millisecond
}

// 1min query data
//   - https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-data
//   - endpoint url - https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m&startTime=1633833600000&endTime=1633833900000&limit=1000
func (api API) GetSpotKline(id string, from, to time.Time, tf Timeframe) ([]market_dto.OhlcRow, error) {
	return api.requestKlineData("https://api.binance.com/api/v3/klines", id, from, to, tf)
}

// https://developers.binance.com/docs/derivatives/coin-margined-futures/market-data/Continuous-Contract-Kline-Candlestick-Data#response-example
func (api API) GetFutureKline(id string, from, to time.Time, tf Timeframe) ([]market_dto.OhlcRow, error) {
	return api.requestKlineData("https://fapi.binance.com/fapi/v1/klines", id, from, to, tf)
}

// Futures and Spot markets have the same data structure. The only difference
// is the endpoint url.
func (api API) requestKlineData(baseUrl string, id string, from, to time.Time, timeframe Timeframe) ([]market_dto.OhlcRow, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	urlSymbol := strings.ToUpper(id)

	// covert time to ms
	t1 := from.UnixMilli()
	t2 := to.UnixMilli()

	const limit = 1000

	url := fmt.Sprintf("%v?symbol=%s&interval=%v&startTime=%d&endTime=%d&limit=%d",
		baseUrl, urlSymbol, timeframe.UrlParam, t1, t2, limit)

	res, err := fetcher.Fetch("GET", url, fetcher.JsonHeader, nil)
	if err != nil {
		return nil, err
	}

	var data [][]any
	if err := encoder.JSON.Unmarshal(res.Body, &data); err != nil {
		return nil, err
	}

	var docs []market_dto.OhlcRow

	for _, d := range data {
		kline, err := parseKineRow(id, d)
		if err != nil {
			return nil, err
		}

		docs = append(docs, *kline)
	}

	return docs, nil
}

// https://developers.binance.com/docs/derivatives/coin-margined-futures/market-data/Continuous-Contract-Kline-Candlestick-Data#response-example
//
// [
//
//			1591258320000,          // Open time
//			"9640.7",               // Open
//			"9642.4",               // High
//			"9640.6",               // Low
//			"9642.0",               // Close (or latest price)
//			"206",                  // Volume
//			1591258379999,          // Close time
//			"2.13660389",           // Base asset volume
//			48,                     // Number of trades
//			"119",                  // Taker buy volume
//			"1.23424865",           // Taker buy base asset volume
//			"0"                     // Ignore.
//	  ]
func parseKineRow(id string, d []any) (*market_dto.OhlcRow, error) {

	if len(d) != 12 {
		return nil, fmt.Errorf("expected 12 fields, got %d", len(d))
	}

	startTime, ok := d[0].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid startTime")
	}
	t1 := time.Unix(int64(startTime/1000), 0)

	open, err := parseFloat(d[1])
	if err != nil {
		return nil, err
	}

	high, err := parseFloat(d[2])
	if err != nil {
		return nil, err
	}

	low, err := parseFloat(d[3])
	if err != nil {
		return nil, err
	}

	close, err := parseFloat(d[4])
	if err != nil {
		return nil, err
	}

	volume, err := parseFloat(d[5])
	if err != nil {
		return nil, err
	}

	baseVol, err := parseFloat(d[7])
	if err != nil {
		return nil, err
	}

	numTrades, ok := d[8].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid numTrades")
	}

	// NOTE: for some reason the timediff between the closeTime and startTime
	// from the api response is not exactly 60 seconds (59999 in ms). That's
	// why we round it to the closest second.
	endTime, ok := d[6].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid endTime")
	}

	t2 := time.Unix(int64((endTime+500)/1000), 0)

	row, err := market_dto.NewOhlcRow(id, t1, t2, open, high, low, close, volume)
	if err != nil {
		return nil, err
	}

	row.SetBaseAssetVolume(baseVol)
	row.SetNumberOfTrades(int64(numTrades))
	return row, nil
}

func parseFloat(val any) (float64, error) {
	switch v := val.(type) {

	case float64:
		return v, nil

	case float32:
		return float64(v), nil

	case int:
		return float64(v), nil

	case string:
		return strconv.ParseFloat(v, 64)

	default:
		return 0, fmt.Errorf("invalid type")
	}
}
