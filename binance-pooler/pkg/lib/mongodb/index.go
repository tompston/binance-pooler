package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexBuilder is a helper for creating indexes for a MongoDB collection
// in a more reusable way.
type IndexBuilder struct {
	indexes []mongo.IndexModel
}

// Indexes returns the indexes that have been added to the IndexBuilder
func (ib *IndexBuilder) Indexes() []mongo.IndexModel { return ib.indexes }

// NewIndexes creates a new IndexBuilder for a given collection.
func NewIndexes() *IndexBuilder { return &IndexBuilder{} }

// Add adds a new index to the IndexBuilder. It supports both single and compound
// indexes. All indexes are created in descending order.
func (ib *IndexBuilder) Add(keys ...string) *IndexBuilder {
	indexKeys := bson.D{}
	for _, key := range keys {
		indexKeys = append(indexKeys, bson.E{Key: key, Value: -1})
	}
	indexModel := mongo.IndexModel{Keys: indexKeys}
	ib.indexes = append(ib.indexes, indexModel)
	return ib
}

// AddUnique is adapted to enforce a unique constraint on a combination of fields.
func (ib *IndexBuilder) AddUnique(keys ...string) *IndexBuilder {
	// Building the compound index keys
	compoundKeys := bson.D{} // Using bson.D to maintain the order of keys
	for _, key := range keys {
		compoundKeys = append(compoundKeys, bson.E{Key: key, Value: -1})
	}

	// Creating the index model with the compound keys
	indexModel := mongo.IndexModel{
		Keys:    compoundKeys,
		Options: options.Index().SetUnique(true),
	}

	// Adding the index model to the indexes slice
	ib.indexes = append(ib.indexes, indexModel)

	return ib
}

// Create creates all the indexes that have been added to the IndexBuilder.
func (ib *IndexBuilder) Create(coll *mongo.Collection) error {
	if ib == nil {
		return fmt.Errorf("ib is nil")
	}

	if coll == nil {
		return fmt.Errorf("coll is nil")
	}

	if len(ib.indexes) == 0 {
		return fmt.Errorf("no indexes to create")
	}

	_, err := coll.Indexes().CreateMany(context.Background(), ib.indexes)
	return err
}

// AvailableIndexes returns all the indexes created for the collection.
func AvailableIndexes(coll *mongo.Collection) ([]bson.M, error) {
	indexesCursor, err := coll.Indexes().List(context.Background())
	if err != nil {
		return nil, err
	}
	defer indexesCursor.Close(context.Background())

	var indexes []bson.M
	err = indexesCursor.All(context.Background(), &indexes)
	return indexes, err
}
