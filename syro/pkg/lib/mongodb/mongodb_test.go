package mongodb

// GO_CONF_PATH="$(pwd)/conf/config.dev.toml" go test -run ^TestIndexCreator$ syro/pkg/settings/db/mongodb -v -count=1
// func TestIndexCreator(t *testing.T) {
// 	db, err := SetupMongdbTest()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer db.Conn().Disconnect(context.Background())

// 	coll := db.coll(TEST_DB, "test_index_creator")

// 	coll.Indexes().DropAll(context.Background())

// 	ib := NewIndexes().
// 		Add("forecast").
// 		Add("area", "reference").
// 		Add("qwe")

// 	if err := ib.Create(coll); err != nil {
// 		t.Errorf("Failed to create indexes: %v", err)
// 	}

// 	availableIndexes, err := AvailableIndexes(coll)
// 	if err != nil {
// 		t.Fatalf("Failed to get indexes: %v", err)
// 	}

// 	if exists, err := IndexExists(coll, "forecast_-1"); err != nil || !exists {
// 		t.Errorf("Index forecast should exist, got: %v, err: %v", exists, err)
// 	}

// 	// Check if the correct number of indexes are created, including the default _id index
// 	if len(availableIndexes) != 4 { //
// 		t.Errorf("Expected 4 indexes, got %d", len(availableIndexes))
// 	}
// }
