package db

import (
	"binance-pooler/pkg/lib/mongodb"
	"binance-pooler/pkg/lib/validate"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// const (
// 	// Default names for the used databases
// 	DEFAULT_NAME = "syro"
// 	// Test db name (all of the collections will be under this db during tests)
// 	TEST_DB = "test"
// )

// Db struct holds the mongodb collection and the schema for the database
type Db struct {
	conn        *mongo.Client
	collections *Collections
	DbName      string
}

// Conn returns the initialized mongodb connection
func (m *Db) Conn() *mongo.Client { return m.conn }

// NewDb returns a new Db struct with the connection to the database and the db schema
func NewDb(uri, dbName string) (*Db, error) {

	if uri == "" {
		return nil, fmt.Errorf("uri for the database is empty")
	}

	if dbName == "" {
		return nil, fmt.Errorf("dbName for the database is empty")
	}

	colls := NewCollections(dbName)

	// Don't proceed if the variable which holds all of the information about the names
	// of the databases and corresponding collections holds an empty string.
	if err := validate.EmptyStringsInStructExist(*colls); err != nil {
		return nil, err
	}

	conn, err := mongodb.New(uri)
	if err != nil {
		return nil, err
	}

	return &Db{conn, colls, dbName}, nil
}

type Collections struct {
	CryptoSpotAsset,
	CryptoSpotOhlc,
	// CryptoFuturesAsset,
	// CryptoFuturesOhlc,
	Logs string
}

// NewDbSchema returns the DbSchema struct which holds the layout of the
// mongodb server databases and collections.
func NewCollections(name string) *Collections {
	return &Collections{
		CryptoSpotAsset: "crypto_spot_asset",
		CryptoSpotOhlc:  "crypto_spot_ohlc",
		Logs:            "logs",
	}
}

// coll returns a mongo collection from the specified db
func (m *Db) coll(dbName, collName string) *mongo.Collection {
	return m.Conn().Database(dbName).Collection(collName)
}

func (m *Db) CryptoSpotAssetColl() *mongo.Collection {
	return m.Conn().Database(m.DbName).Collection(m.collections.CryptoSpotAsset)
}

func (m *Db) CryptoSpotOhlcColl() *mongo.Collection {
	return m.Conn().Database(m.DbName).Collection(m.collections.CryptoSpotOhlc)
}

// Collection to which all logs are written
func (m *Db) LogsCollection() *mongo.Collection {
	return m.coll(m.DbName, m.collections.Logs)
}

// Util function for creating a temporary test collection under the test db
func (m *Db) TestCollection(collName string) *mongo.Collection {
	return m.Conn().Database(m.DbName).Collection(collName)
}
