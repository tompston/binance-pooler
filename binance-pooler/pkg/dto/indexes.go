package dto

import (
	"binance-pooler/pkg/app/db"
	"binance-pooler/pkg/dto/market_dto"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func SetupMongoIndexes(db *db.Db) error {
	assetColls := []*mongo.Collection{
		db.CryptoSpotAssetColl(),
		db.CryptoFuturesAssetColl(),
	}

	for _, coll := range assetColls {
		if err := market_dto.CreateAssetIndexes(coll); err != nil {
			return fmt.Errorf("failed to create indexes for %v: %v", coll.Name(), err)
		}
	}

	olhcColls := []*mongo.Collection{
		db.CryptoSpotOhlcColl(),
		db.CryptoFuturesOhlcColl(),
	}

	for _, coll := range olhcColls {
		if err := market_dto.CreateOhlcIndexes(coll); err != nil {
			return fmt.Errorf("failed to create indexes for %v: %v", coll.Name(), err)
		}
	}

	return nil
}
