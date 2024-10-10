package market_dto

import (
	"context"
	"fmt"
	"syro/pkg/lib/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FuturesAsset struct {
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	// Where the data is coming from
	Source string `json:"source" bson:"source"`
	// Id of the asset (e.g. BTCUSDT)
	ID                    string    `json:"id" bson:"id"`
	ContractType          string    `json:"contract_type" bson:"contract_type"`
	DeliveryDate          time.Time `json:"delivery_date" bson:"delivery_date"`
	OnboardDate           time.Time `json:"onboard_date" bson:"onboard_date"`
	Status                string    `json:"status" bson:"status"`
	MaintMarginPercent    float64   `json:"maint_margin_percent" bson:"maint_margin_percent"`
	RequiredMarginPercent float64   `json:"required_margin_percent" bson:"required_margin_percent"`
	BaseAsset             string    `json:"base_asset" bson:"base_asset"`
	QuoteAsset            string    `json:"quote_asset" bson:"quote_asset"`
	MarginAsset           string    `json:"margin_asset" bson:"margin_asset"`
	UnderlyingType        string    `json:"underlying_type" bson:"underlying_type"`
	TriggerProtect        float64   `json:"trigger_protect" bson:"trigger_protect"`
	LiquidationFee        float64   `json:"liquidation_fee" bson:"liquidation_fee"`
	MarketTakeBound       float64   `json:"market_take_bound" bson:"market_take_bound"`
	MaxMoveOrderLimit     float64   `json:"max_move_order_limit" bson:"max_move_order_limit"`
	OrderTypes            []string  `json:"order_types" bson:"order_types"`
}

func UpsertFuturesAssets(data []FuturesAsset, coll *mongo.Collection) (*mongodb.UpsertLog, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to upsert")
	}

	upsertFn := func(row FuturesAsset) error {
		if row.ID == "" || row.Source == "" {
			return fmt.Errorf("id or source is empty")
		}

		filter := bson.M{"id": row.ID, "source": row.Source}
		_, err := coll.UpdateOne(context.Background(), filter, bson.M{"$set": row}, mongodb.UpsertOpt)
		return err
	}

	start := time.Now()

	for _, row := range data {
		if err := upsertFn(row); err != nil {
			return nil, err
		}
	}

	return mongodb.NewUpsertLog(coll, time.Time{}, time.Time{}, len(data), start), nil
}

func GetFuturesAssetByID(coll *mongo.Collection, id string) (FuturesAsset, error) {
	var doc FuturesAsset
	filter := bson.M{"id": id}
	err := coll.FindOne(context.Background(), filter).Decode(&doc)
	return doc, err
}

func CreateFuturesAssetIndexes(coll *mongo.Collection) error {
	return mongodb.NewIndexes().Add("id").Add("source").Create(coll)
}
