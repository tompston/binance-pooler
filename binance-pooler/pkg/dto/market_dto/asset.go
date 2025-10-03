package market_dto

import (
	"binance-pooler/pkg/lib/mongodb"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Asset[T any] struct {
	AssetBase AssetBase `json:",inline" bson:",inline"` // Fields shared between spot and futures assets
	Data      T         `json:"data" bson:"data"`       // Asset specific data
}

type AssetBase struct {
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	Source     string    `json:"source" bson:"source"` // Where the data is coming from
	Symbol     string    `json:"symbol" bson:"symbol"` // Symbol of the asset (e.g. BTCUSDT)
	Status     string    `json:"status" bson:"status"`
	BaseAsset  string    `json:"base_asset" bson:"base_asset"`
	QuoteAsset string    `json:"quote_asset" bson:"quote_asset"`
	OrderTypes []string  `json:"order_types" bson:"order_types"`
}

func CreateAssetIndexes(coll *mongo.Collection) error {
	return mongodb.NewIndexes().Add("symbol").Add("source", "status").Create(coll)
}

type SpotAsset = Asset[SpotAssetData]
type SpotAssetData struct {
	BaseAssetPrecision         float64 `json:"base_asset_precision" bson:"base_asset_precision"`
	QuotePrecision             float64 `json:"quote_precision" bson:"quote_precision"`
	QuoteAssetPrecision        float64 `json:"quote_asset_precision" bson:"quote_asset_precision"`
	BaseCommissionPrecision    float64 `json:"base_commission_precision" bson:"base_commission_precision"`
	QuoteCommissionPrecision   float64 `json:"quote_commission_precision" bson:"quote_commission_precision"`
	IcebergAllowed             bool    `json:"iceberg_allowed" bson:"iceberg_allowed"`
	OcoAllowed                 bool    `json:"oco_allowed" bson:"oco_allowed"`
	OtoAllowed                 bool    `json:"oto_allowed" bson:"oto_allowed"`
	QuoteOrderQtyMarketAllowed bool    `json:"quote_order_qty_market_allowed" bson:"quote_order_qty_market_allowed"`
	AllowTrailingStop          bool    `json:"allow_trailing_stop" bson:"allow_trailing_stop"`
	CancelReplaceAllowed       bool    `json:"canel_replace_allowed" bson:"cancelRepcanel_replace_allowedlaceAllowed"`
	IsSpotTradingAllowed       bool    `json:"is_spot_trading_allowed" bson:"is_spot_trading_allowed"`
	IsMarginTradingAllowed     bool    `json:"is_margin_trading_allowed" bson:"is_margin_trading_allowed"`
}

type FuturesAsset = Asset[FuturesAssetData]
type FuturesAssetData struct {
	ContractType          string    `json:"contract_type" bson:"contract_type"`
	DeliveryDate          time.Time `json:"delivery_date" bson:"delivery_date"`
	OnboardDate           time.Time `json:"onboard_date" bson:"onboard_date"`
	MaintMarginPercent    float64   `json:"maint_margin_percent" bson:"maint_margin_percent"`
	RequiredMarginPercent float64   `json:"required_margin_percent" bson:"required_margin_percent"`
	MarginAsset           string    `json:"margin_asset" bson:"margin_asset"`
	UnderlyingType        string    `json:"underlying_type" bson:"underlying_type"`
	TriggerProtect        float64   `json:"trigger_protect" bson:"trigger_protect"`
	LiquidationFee        float64   `json:"liquidation_fee" bson:"liquidation_fee"`
	MarketTakeBound       float64   `json:"market_take_bound" bson:"market_take_bound"`
	MaxMoveOrderLimit     float64   `json:"max_move_order_limit" bson:"max_move_order_limit"`
}

func UpsertAssets[T any](data []Asset[T], coll *mongo.Collection) (*mongodb.UpsertLog, error) {
	start := time.Now()

	if len(data) == 0 {
		return nil, fmt.Errorf("no data to upsert")
	}

	upsertFn := func(row Asset[T]) error {
		if row.AssetBase.Symbol == "" || row.AssetBase.Source == "" {
			return fmt.Errorf("symbol or source is empty")
		}

		filter := bson.M{"symbol": row.AssetBase.Symbol, "source": row.AssetBase.Source}
		_, err := coll.UpdateOne(ctx, filter, bson.M{"$set": row}, mongodb.UpsertOpt)
		return err
	}

	for _, row := range data {
		if err := upsertFn(row); err != nil {
			return nil, err
		}
	}

	return mongodb.NewUpsertLog(coll, time.Time{}, time.Time{}, len(data), start), nil
}

func GetAssets(coll *mongo.Collection, filter bson.M, opt *options.FindOptions) ([]AssetBase, error) {
	var docs []AssetBase
	err := mongodb.GetAllDocumentsWithTypes(coll, filter, opt, &docs)
	return docs, err
}
