package dto

import (
	"binance-pooler/pkg/app/db"
	"binance-pooler/pkg/dto/market_dto"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func SetupMongoIndexes(db *db.Db) error {

	mongoInterface := market_dto.NewMongoInterface()

	if err := mongoInterface.CreateAssetIndexes(db.CryptoFuturesAssetColl()); err != nil {
		return fmt.Errorf("failed to create indexes for %v: %v", db.CryptoFuturesAssetColl().Name(), err)
	}

	if err := mongoInterface.CreateAssetIndexes(db.CryptoSpotAssetColl()); err != nil {
		return fmt.Errorf("failed to create indexes for %v: %v", db.CryptoSpotAssetColl().Name(), err)
	}

	olhcColls := []*mongo.Collection{
		db.CryptoSpotOhlcColl(),
		db.CryptoFuturesOhlcColl()}

	for _, coll := range olhcColls {
		if err := mongoInterface.CreateOhlcIndexes(coll); err != nil {
			return fmt.Errorf("failed to create indexes for %v: %v", coll.Name(), err)
		}
	}

	return nil
}
