package binance_service

import (
	"binance-pooler/pkg/providers/binance"
	"context"
	"fmt"

	"github.com/tompston/syro"

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

		log, err := marketdb.UpsertFuturesAssets(docs, coll)
		if err != nil {
			return err
		}

		s.log().Info("upserted binance fututes info", syro.LogFields{"log": log})

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

		log, err := marketdb.UpsertSpotAssets(docs, coll)
		if err != nil {
			return err
		}

		s.log().Info("upserted binance spot info", syro.LogFields{"log": log})

		return nil
	}

	s.log().Info(fmt.Sprintf("futures assets already exist in %v collection, skipping setup", coll.Name()))
	return nil
}
