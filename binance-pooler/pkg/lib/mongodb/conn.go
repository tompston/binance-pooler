// Package mongodb contains functions related to connecting
// and interacting with the mongodb database.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// New returns a new Db struct with the connection to the database and the db schema
func New(uri string) (*mongo.Client, error) {

	if uri == "" {
		return nil, fmt.Errorf("uri for mongodb connection is empty")
	}

	opt := options.Client().
		SetMaxPoolSize(20).                  // Set the maximum number of connections in the connection pool
		SetMaxConnIdleTime(10 * time.Minute) // Close idle connections after the specified time

	opt.ApplyURI(uri)

	conn, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %v", err)
	}

	return conn, nil
}

func Coll(conn *mongo.Client, dbName, collName string) *mongo.Collection {
	return conn.Database(dbName).Collection(collName)
}
