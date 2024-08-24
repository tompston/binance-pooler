package scheduler

import (
	"context"
	"errors"
	"fmt"
	"syro/pkg/lib/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage implementation of the Storage interface
type MongoStorage struct {
	cronListColl    *mongo.Collection
	cronHistoryColl *mongo.Collection
}

// NOTE: add optional auto delete index?
func NewMongoStorage(cronListColl, cronHistoryColl *mongo.Collection) (*MongoStorage, error) {
	if cronListColl == nil || cronHistoryColl == nil {
		return nil, fmt.Errorf("collections cannot be nil")
	}

	// Create indexes for the collections
	if err := mongodb.NewIndexes().
		Add("name").
		Add("status").
		Add("frequency").
		Create(cronListColl); err != nil {
		return nil, err
	}

	// Create indexes for the collections
	if err := mongodb.NewIndexes().
		Add("name").
		Add("initialized_at").
		Add("execution_time").
		Create(cronHistoryColl); err != nil {
		return nil, err
	}

	return &MongoStorage{
		cronListColl:    cronListColl,
		cronHistoryColl: cronHistoryColl,
	}, nil
}

// TODO: refactor so that filter is a variadic parameter
func (m *MongoStorage) AllJobs() ([]JobInfo, error) {
	var docs []JobInfo
	err := mongodb.GetAllDocumentsWithTypes(m.cronListColl, bson.M{}, nil, &docs)
	return docs, err
}

// RegisterJob upsert the job name in the database. If the job does not exist,
// set the created_at field to the current time. If the job already exists,
// update the updated_at field to the current time.
func (m *MongoStorage) RegisterJob(source, name, freq, descr string, status JobStatus, fnErr error) error {
	opt := mongodb.UpsertOpt

	set := bson.M{
		"name":        name,
		"frequency":   freq,
		"status":      status,
		"source":      source,
		"description": descr,
	}

	if fnErr != nil {
		set["exited_with_error"] = true
		set["error"] = fnErr.Error()
	} else {
		set["exited_with_error"] = false
		set["error"] = ""
	}

	_, err := m.cronListColl.UpdateOne(context.Background(), bson.M{"name": name}, bson.M{
		"$set":         set,
		"$setOnInsert": bson.M{"created_at": time.Now().UTC()},
		"$currentDate": bson.M{"updated_at": true},
	}, opt)

	return err
}

// Register the execution of a job in the database
func (m *MongoStorage) RegisterExecution(ex *ExecutionLog) error {
	if ex == nil {
		return fmt.Errorf("job execution cannot be nil")
	}

	_, err := m.cronHistoryColl.InsertOne(context.Background(), ex)
	return err
}

// FindExecutions returns a list of executions based on the filter
func (m *MongoStorage) FindExecutions(filter ExecutionFilter) ([]ExecutionLog, error) {
	queryFilter := bson.M{}

	// if the from and to fields are not zero, add them to the query filter
	if !filter.From.IsZero() && !filter.To.IsZero() {
		if filter.From.After(filter.To) {
			return nil, errors.New("from date cannot be after to date")
		}

		queryFilter["time"] = bson.M{"$gte": filter.From, "$lte": filter.To}
	}

	if filter.ExecutionLog.Source != "" {
		queryFilter["source"] = filter.ExecutionLog.Source
	}

	if filter.ExecutionLog.Name != "" {
		queryFilter["name"] = filter.ExecutionLog.Name
	}

	if filter.ExecutionLog.Error != "" {
		queryFilter["error"] = filter.ExecutionLog.Error
	}

	execTime := filter.ExecutionLog.ExecutionTime
	if execTime != 0 {
		if execTime < 0 {
			return nil, errors.New("execution time cannot be negative")
		}

		// where greater than or equal to the execution time
		queryFilter["execution_time"] = bson.M{"$gte": execTime}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "initialized_at", Value: -1}}).
		SetLimit(filter.Limit).
		SetSkip(filter.Skip)

	var docs []ExecutionLog
	err := mongodb.GetAllDocumentsWithTypes(m.cronHistoryColl, queryFilter, opts, &docs)
	return docs, err
}
