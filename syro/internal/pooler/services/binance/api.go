package binance

import (
	"fmt"
	"strconv"
	"strings"
	"syro/pkg/lib/encoder"
	"syro/pkg/lib/fetcher"
	"syro/pkg/lib/timeset"
	"syro/pkg/models/market_model"
	"time"
)

const SOURCE = "binance"

type API struct{}

func NewAPI() API { return API{} }

type Timeframe struct {
	UrlParam string
	Milis    int64
}

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

const minInMillis = 60 * 1000

var (
	TIMEFRAME_1M  = Timeframe{"1m", 1 * minInMillis}
	TIMEFRAME_5M  = Timeframe{"5m", 5 * minInMillis}
	TIMEFRAME_15M = Timeframe{"15m", 15 * minInMillis}
	TIMEFRAME_30M = Timeframe{"30m", 30 * minInMillis}
	TIMEFRAME_1H  = Timeframe{"1h", 60 * minInMillis}
)

// // 1min query data
// //   - https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-data
// //   - endpoint url - https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m&startTime=1633833600000&endTime=1633833900000&limit=1000
func (api API) GetSpotKline(id string, from, to time.Time, tf Timeframe) ([]market_model.OhlcRow, error) {
	return api.requestKlineData("https://api.binance.com/api/v3/klines", id, from, to, tf)
}

// https://developers.binance.com/docs/derivatives/coin-margined-futures/market-data/Continuous-Contract-Kline-Candlestick-Data#response-example
func (api API) GetFutureKline(id string, from, to time.Time, tf Timeframe) ([]market_model.OhlcRow, error) {
	return api.requestKlineData("https://fapi.binance.com/fapi/v1/klines", id, from, to, tf)
}

// Futures and Spot markets have the same data structure. The only difference
// is the endpoint url.
func (api API) requestKlineData(baseUrl string, id string, from, to time.Time, timeframe Timeframe) ([]market_model.OhlcRow, error) {
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

	var docs []market_model.OhlcRow

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
func parseKineRow(id string, d []any) (*market_model.OhlcRow, error) {

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

	row, err := market_model.NewOhlcRow(id, t1, t2, open, high, low, close, volume)
	if err != nil {
		return nil, err
	}

	row.SetBaseAssetVolume(baseVol)
	row.SetNumberOfTrades(int64(numTrades))
	return row, nil
}

func (api API) GetAllFutureSymbols() ([]market_model.FuturesAsset, error) {
	type Response struct {
		Timezone    string `json:"timezone"`
		ServerTime  int64  `json:"serverTime"`
		FuturesType string `json:"futuresType"`
		RateLimits  []struct {
			RateLimitType string `json:"rateLimitType"`
			Interval      string `json:"interval"`
			IntervalNum   int    `json:"intervalNum"`
			Limit         int    `json:"limit"`
		} `json:"rateLimits"`
		ExchangeFilters []interface{} `json:"exchangeFilters"`
		Assets          []struct {
			Asset             string `json:"asset"`
			MarginAvailable   bool   `json:"marginAvailable"`
			AutoAssetExchange string `json:"autoAssetExchange"`
		} `json:"assets"`
		Symbols []struct {
			Symbol                string   `json:"symbol"`
			Pair                  string   `json:"pair"`
			ContractType          string   `json:"contractType"`
			DeliveryDate          int64    `json:"deliveryDate"`
			OnboardDate           int64    `json:"onboardDate"`
			Status                string   `json:"status"`
			MaintMarginPercent    string   `json:"maintMarginPercent"`
			RequiredMarginPercent string   `json:"requiredMarginPercent"`
			BaseAsset             string   `json:"baseAsset"`
			QuoteAsset            string   `json:"quoteAsset"`
			MarginAsset           string   `json:"marginAsset"`
			PricePrecision        int      `json:"pricePrecision"`
			QuantityPrecision     int      `json:"quantityPrecision"`
			BaseAssetPrecision    int      `json:"baseAssetPrecision"`
			QuotePrecision        int      `json:"quotePrecision"`
			UnderlyingType        string   `json:"underlyingType"`
			UnderlyingSubType     []string `json:"underlyingSubType"`
			SettlePlan            int      `json:"settlePlan"`
			TriggerProtect        string   `json:"triggerProtect"`
			LiquidationFee        string   `json:"liquidationFee"`
			MarketTakeBound       string   `json:"marketTakeBound"`
			MaxMoveOrderLimit     float64  `json:"maxMoveOrderLimit"`
			Filters               []struct {
				FilterType        string `json:"filterType"`
				MaxPrice          string `json:"maxPrice,omitempty"`
				MinPrice          string `json:"minPrice,omitempty"`
				TickSize          string `json:"tickSize,omitempty"`
				StepSize          string `json:"stepSize,omitempty"`
				MinQty            string `json:"minQty,omitempty"`
				MaxQty            string `json:"maxQty,omitempty"`
				Limit             int    `json:"limit,omitempty"`
				Notional          string `json:"notional,omitempty"`
				MultiplierUp      string `json:"multiplierUp,omitempty"`
				MultiplierDown    string `json:"multiplierDown,omitempty"`
				MultiplierDecimal string `json:"multiplierDecimal,omitempty"`
			} `json:"filters"`
			OrderTypes  []string `json:"orderTypes"`
			TimeInForce []string `json:"timeInForce"`
		} `json:"symbols"`
	}

	url := "https://fapi.binance.com/fapi/v1/exchangeInfo"
	res, err := fetcher.Fetch("GET", url, fetcher.JsonHeader, nil)
	if err != nil {
		return nil, err
	}

	var data Response
	if err := encoder.JSON.Unmarshal(res.Body, &data); err != nil {
		return nil, err
	}

	var assets []market_model.FuturesAsset

	for _, symbol := range data.Symbols {
		id := symbol.Symbol
		if id == "" {
			continue
		}

		deliveryDate := timeset.UnixMillisToTime(symbol.DeliveryDate)
		onboardDate := timeset.UnixMillisToTime(symbol.OnboardDate)

		maintMarginPercent, err := parseFloat(symbol.MaintMarginPercent)
		if err != nil {
			continue
		}

		requiredMargin, err := parseFloat(symbol.RequiredMarginPercent)
		if err != nil {
			continue
		}

		triggerProtect, err := parseFloat(symbol.TriggerProtect)
		if err != nil {
			continue
		}

		liquidationFee, err := parseFloat(symbol.LiquidationFee)
		if err != nil {
			continue
		}

		marketTakeBound, err := parseFloat(symbol.MarketTakeBound)
		if err != nil {
			continue
		}

		asset := market_model.FuturesAsset{
			UpdatedAt:             time.Now().UTC(),
			ID:                    id,
			Source:                SOURCE,
			ContractType:          symbol.ContractType,
			DeliveryDate:          deliveryDate,
			OnboardDate:           onboardDate,
			Status:                symbol.Status,
			MaintMarginPercent:    maintMarginPercent,
			RequiredMarginPercent: requiredMargin,
			BaseAsset:             symbol.BaseAsset,
			QuoteAsset:            symbol.QuoteAsset,
			MarginAsset:           symbol.MarginAsset,
			UnderlyingType:        symbol.UnderlyingType,
			TriggerProtect:        triggerProtect,
			LiquidationFee:        liquidationFee,
			MarketTakeBound:       marketTakeBound,
			MaxMoveOrderLimit:     symbol.MaxMoveOrderLimit,
			OrderTypes:            symbol.OrderTypes,
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func parseFloat(v any) (float64, error) {
	// switch v.(type) {

	// case string:
	// 	if f, ok := v.(string); ok {
	// 		return strconv.ParseFloat(f, 64)
	// 	}

	// case float64:
	// 	if f, ok := v.(float64); ok {
	// 		return f, nil
	// 	}

	// default:
	// 	return 0, fmt.Errorf("invalid type")
	// }

	if val, ok := v.(string); ok {
		return strconv.ParseFloat(val, 64)
	}

	if val, ok := v.(float64); ok {
		return val, nil
	}

	return 0, fmt.Errorf("invalid type")
}
