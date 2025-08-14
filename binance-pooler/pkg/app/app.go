package app

import (
	"binance-pooler/pkg/app/db"
	"binance-pooler/pkg/app/settings"
	"binance-pooler/pkg/dto"
	"binance-pooler/pkg/lib/mongodb"
	"context"
	"fmt"
	"testing"

	"github.com/tompston/syro"
)

const (
	VERSION = "0.0.1"
	// Environment variables keys which store the default telegram bot info
	ENV_TELEGRAM_TOKEN_KEY   = "binance-pooler_TG_BOT_TOKEN"
	ENV_TELEGRAM_CHAT_ID_KEY = "binance-pooler_TG_BOT_CHAT_ID"
)

// App holds the state that is needed by the internal packages of the app.
type App struct {
	conf        *settings.TomlConfig
	db          *db.Db
	cronStorage syro.CronStorage
	logger      syro.Logger
}

func (a *App) Conf() *settings.TomlConfig    { return a.conf }
func (a *App) Db() *db.Db                    { return a.db }
func (a *App) CronStorage() syro.CronStorage { return a.cronStorage }
func (a *App) Logger() syro.Logger           { return a.logger }

var Env = &settings.Env{
	DefaultConfigPath: "./conf/config.dev.toml",
	ConfigPathKey:     "GO_CONF_PATH",
	TestModeKey:       "GO_TEST_MODE",
	UseTestDbKey:      "GO_USE_TEST_DB",
	IsProductionKey:   "GO_IS_PRODUCTION",
}

// New returns a new App struct with the specified configuration. If the
// optional debugMode argument is set to true, the app will write all
// of the collections under a single database called "test".
func New(ctx context.Context, debugMode ...bool) (*App, error) {

	confPath := Env.GetConfigPath()

	fmt.Printf(" * initializing app with the %v config...\n", confPath)
	fmt.Printf(" * Is production: %v\n", Env.IsProduction())
	fmt.Printf(" * version: %v\n", VERSION)

	conf, err := settings.NewConfig(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	dbSchema := db.SetDbSchemaBasedOnDebugMode(Env, debugMode...)
	fmt.Printf(" * using db: %v\n", dbSchema.Name)

	mongo := conf.Mongo
	db, err := db.NewDb(mongo.Host, mongo.Port, mongo.Username, mongo.Password, dbSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %v", err)
	}

	if err := dto.SetupMongoEnv(db); err != nil {
		return nil, fmt.Errorf("failed to setup mongodb environment: %v", err)
	}

	cronStorage, err := syro.NewMongoCronStorage(
		mongodb.Coll(db.Conn(), dbSchema.Name, "cron_list"),
		mongodb.Coll(db.Conn(), dbSchema.Name, "cron_history"))
	if err != nil {
		return nil, fmt.Errorf("failed to create cron storage: %v", err)
	}

	logger := syro.NewMongoLogger(db.LogsCollection(), nil)
	if err := logger.CreateIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes for logs collection: %v", err)
	}

	return &App{
		conf:        conf,
		db:          db,
		logger:      logger,
		cronStorage: cronStorage,
	}, nil
}

// Exit closes the MongoDB connection. It should be called with the
// defer keyword after the New function is called.
func (a *App) Exit(ctx context.Context) error {
	if err := a.Db().Conn().Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to close mongodb conn: %v", err)
	}
	return nil
}

// SetupTestEnvironment sets up a test environment for the app.
func SetupTestEnvironment(t *testing.T, useProductionDb ...bool) (*App, func()) {
	debugMode := true
	if len(useProductionDb) == 1 && useProductionDb[0] {
		debugMode = false
	}

	app, err := New(context.Background(), debugMode)
	if err != nil {
		t.Fatalf("failed to setup environment: %v", err)
	}

	cleanup := func() { app.Exit(context.Background()) }
	return app, cleanup
}
