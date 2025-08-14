// Util functions on top of the mongo driver, which don't have a
// reference to the current project.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// The upsert option passed to mongodb query functions
var UpsertOpt = options.Update().SetUpsert(true)

func OrderDescending(field string) *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: field, Value: -1}})
}

func OrderAscending(field string) *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: field, Value: 1}})
}

// TimeRangeFilter returns a filter which matches documents where the given field
// is between the given from and to times.
func TimeRangeFilter(fielName string, from, to time.Time) bson.M {
	return bson.M{fielName: bson.M{"$gte": from.UTC(), "$lte": to.UTC()}}
}

// GetLastDocumentWithTypes queries for a document and returns it, if it exists.
//
// Example
//
//	var row MyType
//	err := mongodb.GetLastDocumentWithTypes(coll, bson.M{"start_date": -1}, &row)
//	return row, err
func GetLastDocumentWithTypes(coll *mongo.Collection, sort primitive.M, filter, results any) error {
	err := coll.FindOne(context.Background(), filter, options.FindOne().SetSort(sort)).Decode(results)
	return err
}

// GetAllDocumentsWithTypes queries for all documents and returns them,
// if they exist. Create an empty slice of the type you want to get
// the results in and pass it as the last parameter.
func GetAllDocumentsWithTypes(coll *mongo.Collection, filter primitive.M, options *options.FindOptions, results any) error {
	ctx := context.Background()
	cur, err := coll.Find(ctx, filter, options)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return cur.All(ctx, results)
}

func GetDocumentWithTypes(coll *mongo.Collection, filter primitive.M, options *options.FindOneOptions, results any) error {
	err := coll.FindOne(context.Background(), filter, options).Decode(results)
	return err
}

// DeleteField deletes the specified field from all documents in the collection.
func DeleteField(coll *mongo.Collection, fieldName string) error {
	if fieldName == "" {
		return fmt.Errorf("field name is empty")
	}

	update := bson.M{"$unset": bson.M{fieldName: ""}}

	// Perform an update many operation to apply the update to all documents
	if _, err := coll.UpdateMany(context.Background(), bson.M{}, update); err != nil {
		return fmt.Errorf("failed to delete '%v' field: %v", fieldName, err)
	}

	fmt.Printf("Successfully deleted '%v' field from all documents.\n", fieldName)
	return nil
}

func DeleteIndex(coll *mongo.Collection, indexName string) error {
	_, err := coll.Indexes().DropOne(context.Background(), indexName)
	return err
}

// QueryParams holds the parameters that are used to query the database for documents.
type QueryParams struct {
	Coll    *mongo.Collection
	Filter  primitive.M
	Options *options.FindOptions
}

// GetDocuments is a helper function that queries the database for documents
// and returns them as types of the data argument.
func GetDocuments[T any](params QueryParams, data *[]T) error {
	return GetAllDocumentsWithTypes(params.Coll, params.Filter, params.Options, data)
}

func DeleteByID(coll *mongo.Collection, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = coll.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}

// UpsertLog struct holds meta information about how long it took to upsert documents into a collection.
type UpsertLog struct {
	CollectionName  string    `json:"collection_name" bson:"collection_name"`         // name of the collection into which the rows were inserted
	DbName          string    `json:"db_name" bson:"db_name"`                         // name of the database into which the rows were inserted
	FirstStartTime  time.Time `json:"first_start_time" bson:"first_start_time"`       // first start time of the rows inserted
	LastStartTime   time.Time `json:"last_start_time" bson:"last_start_time"`         // last start time of the rows inserted
	NumUpsertedRows int       `json:"inserted_rows_count" bson:"inserted_rows_count"` // number of rows inserted
	ElapsedTime     float64   `json:"elapsed_time" bson:"elapsed_time"`               // total time it took to upsert the rows (in seconds)
}

// NewLog returns a new Log instance which holds data about the upsert operation
func NewUpsertLog(coll *mongo.Collection, firstStartTime, lastStartTime time.Time, numUpsertedRows int, operationDuration time.Time) *UpsertLog {
	return &UpsertLog{
		DbName:          coll.Database().Name(),
		CollectionName:  coll.Name(),
		FirstStartTime:  firstStartTime,
		LastStartTime:   lastStartTime,
		NumUpsertedRows: numUpsertedRows,
		ElapsedTime:     time.Since(operationDuration).Seconds(),
	}
}

// String returns a string representation of the Log struct
func (l *UpsertLog) String() string {
	if l == nil {
		return "<nil>"
	}

	format := "2006-01-02 15:04:05"

	destination := l.DbName + "." + l.CollectionName

	return fmt.Sprintf(
		"upserted %v rows in the %v coll from the period of %v to %v in %.2f sec",
		l.NumUpsertedRows, destination, l.FirstStartTime.Format(format), l.LastStartTime.Format(format), l.ElapsedTime,
	)
}
