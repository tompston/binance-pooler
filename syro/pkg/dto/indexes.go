package dto

import (
	"fmt"
	"syro/pkg/app/db"
	"syro/pkg/dto/market_dto"
	"syro/pkg/lib/sy"

	"go.mongodb.org/mongo-driver/mongo"
)

// SetupMongoEnv creates the timeseries collection for the logs collection and
// indexes for collections
func SetupMongoEnv(db *db.Db) error {

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

	if err := sy.CreateMongoIndexes(db.LogsCollection()); err != nil {
		return fmt.Errorf("failed to create indexes for %v: %v", db.LogsCollection().Name(), err)
	}

	return nil
}
