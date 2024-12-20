package binance

import (
	"binance-pooler/pkg/dto/market_dto"
	"binance-pooler/pkg/syro/fetcher"
	"binance-pooler/pkg/syro/timeset"
	"encoding/json"
	"time"
)

func (api API) GetAllFutureSymbols() ([]market_dto.FuturesAsset, error) {
	type apiResponse struct {
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
			OrderTypes            []string `json:"orderTypes"`
			TimeInForce           []string `json:"timeInForce"`
			// Filters               []struct {
			// 	FilterType        string `json:"filterType"`
			// 	MaxPrice          string `json:"maxPrice,omitempty"`
			// 	MinPrice          string `json:"minPrice,omitempty"`
			// 	TickSize          string `json:"tickSize,omitempty"`
			// 	StepSize          string `json:"stepSize,omitempty"`
			// 	MinQty            string `json:"minQty,omitempty"`
			// 	MaxQty            string `json:"maxQty,omitempty"`
			// 	Limit             int    `json:"limit,omitempty"`
			// 	Notional          string `json:"notional,omitempty"`
			// 	MultiplierUp      string `json:"multiplierUp,omitempty"`
			// 	MultiplierDown    string `json:"multiplierDown,omitempty"`
			// 	MultiplierDecimal string `json:"multiplierDecimal,omitempty"`
			// } `json:"filters"`
		} `json:"symbols"`
	}

	res, err := fetcher.Fetch("GET", "https://fapi.binance.com/fapi/v1/exchangeInfo", fetcher.JsonHeader)
	if err != nil {
		return nil, err
	}

	var data apiResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return nil, err
	}

	var assets []market_dto.FuturesAsset

	for _, symbol := range data.Symbols {
		symb := symbol.Symbol
		if symb == "" {
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

		asset := market_dto.FuturesAsset{
			UpdatedAt:             time.Now().UTC(),
			Symbol:                symb,
			Source:                Source,
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

func (api API) GetAllSpotAssets() ([]market_dto.SpotAsset, error) {
	type apiResponse struct {
		Timezone   string `json:"timezone"`
		ServerTime int64  `json:"serverTime"`
		RateLimits []struct {
			RateLimitType string `json:"rateLimitType"`
			Interval      string `json:"interval"`
			IntervalNum   int    `json:"intervalNum"`
			Limit         int    `json:"limit"`
		} `json:"rateLimits"`
		ExchangeFilters []interface{} `json:"exchangeFilters"`
		Symbols         []struct {
			Symbol                          string        `json:"symbol"`
			Status                          string        `json:"status"`
			BaseAsset                       string        `json:"baseAsset"`
			QuoteAsset                      string        `json:"quoteAsset"`
			BaseAssetPrecision              float64       `json:"baseAssetPrecision"`
			QuotePrecision                  float64       `json:"quotePrecision"`
			QuoteAssetPrecision             float64       `json:"quoteAssetPrecision"`
			BaseCommissionPrecision         float64       `json:"baseCommissionPrecision"`
			QuoteCommissionPrecision        float64       `json:"quoteCommissionPrecision"`
			OrderTypes                      []string      `json:"orderTypes"`
			IcebergAllowed                  bool          `json:"icebergAllowed"`
			OcoAllowed                      bool          `json:"ocoAllowed"`
			OtoAllowed                      bool          `json:"otoAllowed"`
			QuoteOrderQtyMarketAllowed      bool          `json:"quoteOrderQtyMarketAllowed"`
			AllowTrailingStop               bool          `json:"allowTrailingStop"`
			CancelReplaceAllowed            bool          `json:"cancelReplaceAllowed"`
			IsSpotTradingAllowed            bool          `json:"isSpotTradingAllowed"`
			IsMarginTradingAllowed          bool          `json:"isMarginTradingAllowed"`
			Permissions                     []interface{} `json:"permissions"`
			PermissionSets                  [][]string    `json:"permissionSets"`
			DefaultSelfTradePreventionMode  string        `json:"defaultSelfTradePreventionMode"`
			AllowedSelfTradePreventionModes []string      `json:"allowedSelfTradePreventionModes"`
			// Filters                    []struct {
			// 	FilterType            string `json:"filterType"`
			// 	MinPrice              string `json:"minPrice,omitempty"`
			// 	MaxPrice              string `json:"maxPrice,omitempty"`
			// 	TickSize              string `json:"tickSize,omitempty"`
			// 	MinQty                string `json:"minQty,omitempty"`
			// 	MaxQty                string `json:"maxQty,omitempty"`
			// 	StepSize              string `json:"stepSize,omitempty"`
			// 	Limit                 int    `json:"limit,omitempty"`
			// 	MinTrailingAboveDelta int    `json:"minTrailingAboveDelta,omitempty"`
			// 	MaxTrailingAboveDelta int    `json:"maxTrailingAboveDelta,omitempty"`
			// 	MinTrailingBelowDelta int    `json:"minTrailingBelowDelta,omitempty"`
			// 	MaxTrailingBelowDelta int    `json:"maxTrailingBelowDelta,omitempty"`
			// 	BidMultiplierUp       string `json:"bidMultiplierUp,omitempty"`
			// 	BidMultiplierDown     string `json:"bidMultiplierDown,omitempty"`
			// 	AskMultiplierUp       string `json:"askMultiplierUp,omitempty"`
			// 	AskMultiplierDown     string `json:"askMultiplierDown,omitempty"`
			// 	AvgPriceMins          int    `json:"avgPriceMins,omitempty"`
			// 	MinNotional           string `json:"minNotional,omitempty"`
			// 	ApplyMinToMarket      bool   `json:"applyMinToMarket,omitempty"`
			// 	MaxNotional           string `json:"maxNotional,omitempty"`
			// 	ApplyMaxToMarket      bool   `json:"applyMaxToMarket,omitempty"`
			// 	MaxNumOrders          int    `json:"maxNumOrders,omitempty"`
			// 	MaxNumAlgoOrders      int    `json:"maxNumAlgoOrders,omitempty"`
			// } `json:"filters"`
		} `json:"symbols"`
	}

	res, err := fetcher.Fetch("GET", "https://api.binance.com/api/v3/exchangeInfo", fetcher.JsonHeader)
	if err != nil {
		return nil, err
	}

	var data apiResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return nil, err
	}

	var docs []market_dto.SpotAsset
	for _, symbol := range data.Symbols {
		symb := symbol.Symbol
		if symb == "" {
			continue
		}

		asset := market_dto.SpotAsset{
			UpdatedAt:                  time.Now().UTC(),
			Source:                     Source,
			Symbol:                     symb,
			Status:                     symbol.Status,
			BaseAsset:                  symbol.BaseAsset,
			QuoteAsset:                 symbol.QuoteAsset,
			BaseAssetPrecision:         symbol.BaseAssetPrecision,
			QuotePrecision:             symbol.QuotePrecision,
			QuoteAssetPrecision:        symbol.QuoteAssetPrecision,
			BaseCommissionPrecision:    symbol.BaseCommissionPrecision,
			QuoteCommissionPrecision:   symbol.QuoteCommissionPrecision,
			OrderTypes:                 symbol.OrderTypes,
			IcebergAllowed:             symbol.IcebergAllowed,
			OcoAllowed:                 symbol.OcoAllowed,
			OtoAllowed:                 symbol.OtoAllowed,
			QuoteOrderQtyMarketAllowed: symbol.QuoteOrderQtyMarketAllowed,
			AllowTrailingStop:          symbol.AllowTrailingStop,
			CancelReplaceAllowed:       symbol.CancelReplaceAllowed,
			IsSpotTradingAllowed:       symbol.IsSpotTradingAllowed,
			IsMarginTradingAllowed:     symbol.IsMarginTradingAllowed,
		}

		docs = append(docs, asset)
	}

	return docs, nil
}
