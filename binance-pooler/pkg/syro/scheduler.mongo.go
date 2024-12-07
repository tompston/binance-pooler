package syro

import (
	"binance-pooler/pkg/lib/mongodb"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage implementation of the Storage interface
type MongoCronStorage struct {
	Options         CronStorageOptions
	cronListColl    *mongo.Collection
	cronHistoryColl *mongo.Collection
}

// NOTE: add optional auto delete index?
func NewMongoCronStorage(cronListColl, cronHistoryColl *mongo.Collection) (*MongoCronStorage, error) {
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

	return &MongoCronStorage{
		cronListColl:    cronListColl,
		cronHistoryColl: cronHistoryColl,
	}, nil
}

func (m *MongoCronStorage) SetOptions(opt CronStorageOptions) CronStorage {
	m.Options = opt
	return m
}

func (m *MongoCronStorage) GetStorageOptions() CronStorageOptions {
	return m.Options
}

// TODO: refactor so that filter is a variadic parameter
func (m *MongoCronStorage) FindJobs() ([]CronInfo, error) {
	var docs []CronInfo
	err := mongodb.GetAllDocumentsWithTypes(m.cronListColl, bson.M{}, nil, &docs)
	return docs, err
}

// TODO: test this function
func (m *MongoCronStorage) SetJobsToInactive(source string) error {
	filter := bson.M{"source": source}
	update := bson.M{"$set": bson.M{"status": JobStatusInactive}}
	_, err := m.cronListColl.UpdateMany(context.Background(), filter, update)
	return err
}

// RegisterJob upsert the job name in the database based on the source
// and the job name. If the job does not exist, set the created_at
// field to the current time. If the job already exists,
// update the updated_at field to the current time.
func (m *MongoCronStorage) RegisterJob(source, name, freq, descr string, status JobStatus, fnErr error) error {
	filter := bson.M{
		"source": source,
		"name":   name,
	}

	set := bson.M{
		"frequency":   freq,
		"status":      status,
		"description": descr,
		"updated_at":  time.Now().UTC(),
	}

	if fnErr != nil {
		set["exited_with_error"] = true
		set["error"] = fnErr.Error()
	} else {
		set["exited_with_error"] = false
		set["error"] = ""
	}

	_, err := m.cronListColl.UpdateOne(context.Background(), filter, bson.M{
		"$set":         set,
		"$setOnInsert": bson.M{"created_at": time.Now().UTC()},
	}, mongodb.UpsertOpt)

	return err
}

// Register the execution of a job in the database
func (m *MongoCronStorage) RegisterExecution(ex *CronExecLog) error {
	if ex == nil {
		return fmt.Errorf("job execution cannot be nil")
	}

	_, err := m.cronHistoryColl.InsertOne(context.Background(), ex)
	return err
}

// FindExecutions returns a list of executions based on the filter
func (m *MongoCronStorage) FindExecutions(filter CronExecFilter) ([]CronExecLog, error) {
	queryFilter := bson.M{}

	// if the from and to fields are not zero, add them to the query filter
	if !filter.From.IsZero() && !filter.To.IsZero() {
		if filter.From.After(filter.To) {
			return nil, errors.New("from date cannot be after to date")
		}

		queryFilter["time"] = bson.M{"$gte": filter.From, "$lte": filter.To}
	}

	if filter.Source != "" {
		queryFilter["source"] = filter.Source
	}

	if filter.Name != "" {
		queryFilter["name"] = filter.Name
	}

	// if filter.Error != "" {
	// 	queryFilter["error"] = filter.Error
	// }

	// execTime := filter.ExecutionTime
	// if execTime != 0 {
	// 	if execTime < 0 {
	// 		return nil, errors.New("execution time cannot be negative")
	// 	}

	// 	// where greater than or equal to the execution time
	// 	queryFilter["execution_time"] = bson.M{"$gte": execTime}
	// }

	opts := options.Find().
		SetSort(bson.D{{Key: "initialized_at", Value: -1}}).
		SetLimit(filter.Limit).
		SetSkip(filter.Skip)

	var docs []CronExecLog
	err := mongodb.GetAllDocumentsWithTypes(m.cronHistoryColl, queryFilter, opts, &docs)
	return docs, err
}
