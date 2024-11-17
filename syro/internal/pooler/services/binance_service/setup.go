package binance_service

import (
	"context"
	"fmt"
	"syro/pkg/dto/market_dto"
	"syro/pkg/providers/binance"
	"syro/pkg/sy"

	"go.mongodb.org/mongo-driver/bson"
)

// If the db does not have any futures assets, scrape the list from the binance api
func (s *service) setupFuturesAssets() error {
	coll := s.app.Db().CryptoFuturesAssetColl()

	filter := bson.M{"source": binance.Source}
	count, err := coll.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}

	if count == 0 {
		s.log().Info("no futures assets found, scraping data")

		docs, err := s.api.GetAllFutureSymbols()
		if err != nil {
			return err
		}

		log, err := market_dto.NewMongoInterface().UpsertFuturesAssets(docs, coll)
		if err != nil {
			return err
		}

		s.log().Info("upserted binance fututes info", sy.LogFields{"log": log})

		return nil
	}

	s.log().Info(fmt.Sprintf("futures assets already exist in %v collection, skipping setup", coll.Name()))
	return nil
}

func (s *service) setupSpotAssets() error {
	coll := s.app.Db().CryptoSpotAssetColl()

	filter := bson.M{"source": binance.Source}
	count, err := coll.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}

	if count == 0 {
		s.log().Info("no spot assets found, scraping data")
		docs, err := s.api.GetAllSpotAssets()
		if err != nil {
			return err
		}

		log, err := market_dto.NewMongoInterface().UpsertSpotAssets(docs, coll)
		if err != nil {
			return err
		}

		s.log().Info("upserted binance spot info", sy.LogFields{"log": log})

		return nil
	}

	s.log().Info(fmt.Sprintf("futures assets already exist in %v collection, skipping setup", coll.Name()))
	return nil
}
