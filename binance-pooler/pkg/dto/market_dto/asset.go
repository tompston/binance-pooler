package market_dto

import (
	"binance-pooler/pkg/lib/mongodb"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (*Mongo) CreateAssetIndexes(coll *mongo.Collection) error {
	return mongodb.NewIndexes().Add("symbol").Add("source", "status").Create(coll)
}

type SpotAsset struct {
	UpdatedAt                  time.Time `json:"updated_at" bson:"updated_at"`
	Source                     string    `json:"source" bson:"source"` // Where the data is coming from
	Symbol                     string    `json:"symbol" bson:"symbol"` // symb of the asset (e.g. BTCUSDT)
	Status                     string    `json:"status" bson:"status"`
	BaseAsset                  string    `json:"base_asset" bson:"base_asset"`
	QuoteAsset                 string    `json:"quote_asset" bson:"quote_asset"`
	BaseAssetPrecision         float64   `json:"base_asset_precision" bson:"base_asset_precision"`
	QuotePrecision             float64   `json:"quote_precision" bson:"quote_precision"`
	QuoteAssetPrecision        float64   `json:"quote_asset_precision" bson:"quote_asset_precision"`
	BaseCommissionPrecision    float64   `json:"base_commission_precision" bson:"base_commission_precision"`
	QuoteCommissionPrecision   float64   `json:"quote_commission_precision" bson:"quote_commission_precision"`
	OrderTypes                 []string  `json:"order_types" bson:"order_types"`
	IcebergAllowed             bool      `json:"iceberg_allowed" bson:"iceberg_allowed"`
	OcoAllowed                 bool      `json:"oco_allowed" bson:"oco_allowed"`
	OtoAllowed                 bool      `json:"oto_allowed" bson:"oto_allowed"`
	QuoteOrderQtyMarketAllowed bool      `json:"quote_order_qty_market_allowed" bson:"quote_order_qty_market_allowed"`
	AllowTrailingStop          bool      `json:"allow_trailing_stop" bson:"allow_trailing_stop"`
	CancelReplaceAllowed       bool      `json:"canel_replace_allowed" bson:"cancelRepcanel_replace_allowedlaceAllowed"`
	IsSpotTradingAllowed       bool      `json:"is_spot_trading_allowed" bson:"is_spot_trading_allowed"`
	IsMarginTradingAllowed     bool      `json:"is_margin_trading_allowed" bson:"is_margin_trading_allowed"`
}

func (*Mongo) UpsertSpotAssets(data []SpotAsset, coll *mongo.Collection) (*mongodb.UpsertLog, error) {
	start := time.Now()

	if len(data) == 0 {
		return nil, fmt.Errorf("no data to upsert")
	}

	upsertFn := func(row SpotAsset) error {
		if row.Symbol == "" || row.Source == "" {
			return fmt.Errorf("symbol or source is empty")
		}

		filter := bson.M{"symbol": row.Symbol, "source": row.Source}
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

func (*Mongo) GetSpotAssets(coll *mongo.Collection, filter bson.M, opt *options.FindOptions) ([]SpotAsset, error) {
	var docs []SpotAsset
	err := mongodb.GetAllDocumentsWithTypes(coll, filter, opt, &docs)
	return docs, err
}
