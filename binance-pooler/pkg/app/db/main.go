package db

import (
	"binance-pooler/pkg/app/settings"
	"binance-pooler/pkg/lib/mongodb"
	"binance-pooler/pkg/lib/validate"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// Default names for the used databases
	DEFAULT_NAME = "syro"
	// Test db name (all of the collections will be under this db during tests)
	TEST_DB = "test"
)

// Db struct holds the mongodb collection and the schema for the database
type Db struct {
	conn   *mongo.Client
	schema *DbSchema
}

// Conn returns the initialized mongodb connection
func (m *Db) Conn() *mongo.Client { return m.conn }

// NewDb returns a new Db struct with the connection to the database and the db schema
func NewDb(host string, port int, username, password string, schema *DbSchema) (*Db, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema for the database is nil")
	}

	// Don't proceed if the variable which holds all of the information about the names
	// of the databases and corresponding collections holds an empty string.
	if err := validate.EmptyStringsInStructExist(*schema); err != nil {
		return nil, err
	}

	conn, err := mongodb.New(host, port, username, password)
	if err != nil {
		return nil, err
	}

	return &Db{conn, schema}, nil
}

func SetupMongdbTest(env *settings.Env, useProdDb ...bool) (*Db, error) {
	if env == nil {
		return nil, fmt.Errorf("env is nil")
	}

	var dbSchema *DbSchema
	if len(useProdDb) == 1 && useProdDb[0] {
		dbSchema = DefaultDbSchema()
	} else {
		dbSchema = TestsDbSchema()
	}

	confPath := env.GetConfigPath()
	conf, err := settings.NewConfig(confPath)
	if err != nil {
		return nil, err
	}

	mongo := conf.Mongo

	return NewDb(mongo.Host, mongo.Port, mongo.Username, mongo.Password, dbSchema)
}

// DbSchema holds a map of all of the names for the dbs and their corresponding collections
type DbSchema struct {
	Name  string
	colls Collections
}

type Collections struct {
	CryptoSpotAsset,
	CryptoSpotOhlc,
	CryptoFuturesAsset,
	CryptoFuturesOhlc,
	Logs string
}

// TestsDbSchema returns the db schema that is set while running tests to isolate
// the test data from the production data
func TestsDbSchema() *DbSchema { return NewDbSchema(TEST_DB) }

// DefaultDbSchema sets the db names that are used when running the app locally or
// in the production
func DefaultDbSchema() *DbSchema { return NewDbSchema(DEFAULT_NAME) }

// Else the default db schema is used.
func SetDbSchemaBasedOnDebugMode(env *settings.Env, debugMode ...bool) *DbSchema {
	if env.ShouldUseTestDb() {
		return TestsDbSchema()
	}

	if len(debugMode) == 1 && debugMode[0] {
		return TestsDbSchema()
	}

	return DefaultDbSchema()
}

// NewDbSchema returns the DbSchema struct which holds the layout of the
// mongodb server databases and collections.
func NewDbSchema(name string) *DbSchema {
	return &DbSchema{
		Name: name,
		colls: Collections{
			CryptoSpotAsset:    "crypto_spot_asset",
			CryptoSpotOhlc:     "crypto_spot_ohlc",
			CryptoFuturesAsset: "crypto_futures_asset",
			CryptoFuturesOhlc:  "crypto_futures_ohlc",
			Logs:               "logs",
		},
	}
}

// coll returns a mongo collection from the specified db
func (m *Db) coll(dbName, collName string) *mongo.Collection {
	return m.Conn().Database(dbName).Collection(collName)
}

func (m *Db) CryptoSpotAssetColl() *mongo.Collection {
	return m.Conn().Database(m.schema.Name).Collection(m.schema.colls.CryptoSpotAsset)
}

func (m *Db) CryptoSpotOhlcColl() *mongo.Collection {
	return m.Conn().Database(m.schema.Name).Collection(m.schema.colls.CryptoSpotOhlc)
}

func (m *Db) CryptoFuturesAssetColl() *mongo.Collection {
	return m.Conn().Database(m.schema.Name).Collection(m.schema.colls.CryptoFuturesAsset)
}

func (m *Db) CryptoFuturesOhlcColl() *mongo.Collection {
	return m.Conn().Database(m.schema.Name).Collection(m.schema.colls.CryptoFuturesOhlc)
}

// Collection to which all logs are written
func (m *Db) LogsCollection() *mongo.Collection {
	return m.coll(m.schema.Name, m.schema.colls.Logs)
}

// Util function for creating a temporary test collection under the test db
func (m *Db) TestCollection(collName string) *mongo.Collection {
	return m.Conn().Database(TEST_DB).Collection(collName)
}
