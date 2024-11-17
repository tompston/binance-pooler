package yahoo_service

import (
	"encoding/json"
	"fmt"
	"syro/pkg/sy/fetcher"
	"time"
)

type RequestInterval string

const (
	Interval1m  RequestInterval = "1m"
	Interval5m  RequestInterval = "5m"
	Interval15m RequestInterval = "15m"
	Interval30m RequestInterval = "30m"
	Interval1h  RequestInterval = "1h"
)

//	curl 'https://query1.finance.yahoo.com/v8/finance/chart/ASM.AS?period1=1722498000&period2=1723564800&interval=5m' \
//	  -H 'accept: */*' \
//	  -H 'accept-language: en-GB,en;q=0.6' \
//	  -H 'priority: u=1, i' \
//	  -H 'sec-ch-ua-mobile: ?0' \
//	  -H 'sec-fetch-dest: empty' \
//	  -H 'sec-fetch-mode: cors' \
//	  -H 'sec-fetch-site: same-site' \
//	  -H 'sec-gpc: 1' -o "qwe.json"
func GetStockData(stockSymbol string, from, to time.Time, interval RequestInterval) error {

	headers := map[string]string{
		"accept":           "*/*",
		"accept-language":  "en-GB,en;q=0.6",
		"priority":         "u=1, i",
		"sec-ch-ua-mobile": "?0",
		"sec-fetch-dest":   "empty",
		"sec-fetch-mode":   "cors",
		"sec-fetch-site":   "same-site",
		"sec-gpc":          "1",
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s",
		stockSymbol, from.Unix(), to.Unix(), interval)

	type apiResponse struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency             string  `json:"currency"`
					Symbol               string  `json:"symbol"`
					ExchangeName         string  `json:"exchangeName"`
					FullExchangeName     string  `json:"fullExchangeName"`
					InstrumentType       string  `json:"instrumentType"`
					FirstTradeDate       int     `json:"firstTradeDate"`
					RegularMarketTime    int     `json:"regularMarketTime"`
					HasPrePostMarketData bool    `json:"hasPrePostMarketData"`
					Gmtoffset            int     `json:"gmtoffset"`
					Timezone             string  `json:"timezone"`
					ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
					RegularMarketPrice   float64 `json:"regularMarketPrice"`
					FiftyTwoWeekHigh     float64 `json:"fiftyTwoWeekHigh"`
					FiftyTwoWeekLow      float64 `json:"fiftyTwoWeekLow"`
					RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
					RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
					RegularMarketVolume  int     `json:"regularMarketVolume"`
					LongName             string  `json:"longName"`
					ShortName            string  `json:"shortName"`
					ChartPreviousClose   float64 `json:"chartPreviousClose"`
					PreviousClose        float64 `json:"previousClose"`
					Scale                int     `json:"scale"`
					PriceHint            int     `json:"priceHint"`
					CurrentTradingPeriod struct {
						Pre struct {
							Timezone  string `json:"timezone"`
							End       int    `json:"end"`
							Start     int    `json:"start"`
							Gmtoffset int    `json:"gmtoffset"`
						} `json:"pre"`
						Regular struct {
							Timezone  string `json:"timezone"`
							End       int    `json:"end"`
							Start     int    `json:"start"`
							Gmtoffset int    `json:"gmtoffset"`
						} `json:"regular"`
						Post struct {
							Timezone  string `json:"timezone"`
							End       int    `json:"end"`
							Start     int    `json:"start"`
							Gmtoffset int    `json:"gmtoffset"`
						} `json:"post"`
					} `json:"currentTradingPeriod"`
					TradingPeriods [][]struct {
						Timezone  string `json:"timezone"`
						End       int    `json:"end"`
						Start     int    `json:"start"`
						Gmtoffset int    `json:"gmtoffset"`
					} `json:"tradingPeriods"`
					DataGranularity string   `json:"dataGranularity"`
					Range           string   `json:"range"`
					ValidRanges     []string `json:"validRanges"`
				} `json:"meta"`
				Timestamp  []int `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Low    []interface{} `json:"low"`
						Open   []interface{} `json:"open"`
						High   []interface{} `json:"high"`
						Volume []interface{} `json:"volume"`
						Close  []interface{} `json:"close"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error interface{} `json:"error"`
		} `json:"chart"`
	}

	res, err := fetcher.Fetch("GET", url, headers)
	if err != nil {
		return err
	}

	var data apiResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return err
	}

	for _, result := range data.Chart.Result {
		for _, timestamp := range result.Timestamp {
			// unix int to time.Time
			t := time.Unix(int64(timestamp), 0)

			fmt.Printf("Timestamp: %v\n", t.UTC())
		}
	}

	return nil
}
