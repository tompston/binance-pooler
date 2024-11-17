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
func New(host string, port int, username, password string) (*mongo.Client, error) {
	opt := options.Client().
		SetMaxPoolSize(20).                  // Set the maximum number of connections in the connection pool
		SetMaxConnIdleTime(10 * time.Minute) // Close idle connections after the specified time

	// If both the username and password exists, use it as the credentials. Else use the non-authenticated url.
	var url string
	if username != "" && password != "" {
		opt.SetAuth(options.Credential{Username: username, Password: password})
		url = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, host, port)
	} else {
		url = fmt.Sprintf("mongodb://%s:%d", host, port)
	}

	opt.ApplyURI(url)

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
