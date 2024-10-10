package dto

import (
	"fmt"
	"syro/pkg/app/db"
	"syro/pkg/dto/market_dto"
	"syro/pkg/lib/logger"

	"go.mongodb.org/mongo-driver/mongo"
)

// SetupMongoEnv creates the timeseries collection for the logs collection and
// indexes for collections
func SetupMongoEnv(db *db.Db) error {

	if err := market_dto.CreateFuturesAssetIndexes(db.CryptoFuturesAssetColl()); err != nil {
		return fmt.Errorf("failed to create indexes for %v: %v", db.CryptoFuturesAssetColl().Name(), err)
	}

	olhcColls := []*mongo.Collection{
		db.CryptoSpotOhlcColl(),
		db.CryptoFuturesOhlcColl()}

	for _, coll := range olhcColls {
		if err := market_dto.CreateOhlcIndexes(coll); err != nil {
			return fmt.Errorf("failed to create indexes for %v: %v", coll.Name(), err)
		}
	}

	if err := logger.CreateMongoIndexes(db.LogsCollection()); err != nil {
		return fmt.Errorf("failed to create indexes for %v: %v", db.LogsCollection().Name(), err)
	}

	return nil
}
