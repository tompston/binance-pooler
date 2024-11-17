package sy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"syro/pkg/lib/mongodb"
	"syro/pkg/lib/utils"
	"syro/pkg/lib/validate"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestLog(t *testing.T) {

	t.Run("test log creation", func(t *testing.T) {
		// logger := NewConsoleLogger(nil)

		log := newLog(ERROR, "qweqwe", "my-source", "my-event", "my-event-id")

		decoded, err := utils.DecodeStructToStrings(log)
		if err != nil {
			t.Error(err)
		}

		// parse the created_at field from the json string and check it the time is
		// within the last 2 seconds
		type parsed struct {
			CreatedAt time.Time `json:"time" bson:"time"`
		}

		t.Run("test json unmarshalling", func(t *testing.T) {
			if err := validate.StringIncludes(decoded.JSON, []string{
				`"level":4`,
				`message":"qweqwe"`,
				`"source":"my-source"`,
				`"event":"my-event"`,
				`"event_id":"my-event-id"`,
				`"time":`,
			}); err != nil {
				t.Fatal(err)
			}

			var v parsed
			if err := json.Unmarshal([]byte(decoded.JSON), &v); err != nil {
				t.Error(err)
			}

			if v.CreatedAt.Before(time.Now().Add(-2 * time.Second)) {
				t.Error("The time time is not within the last 2 seconds")
			}

			// Check the timezone of the created_at field
			if v.CreatedAt.Location().String() != "UTC" {
				t.Error("The created_at time is not in UTC")
			}
		})

		t.Run("test bson unmarshalling", func(t *testing.T) {
			if err := validate.StringIncludes(decoded.BSON, []string{
				`"time":{"$date":`,
				`message":"qweqwe"`,
				`"source":"my-source"`,
				`"event":"my-event"`,
				`"event_id":"my-event-id"`,
			}); err != nil {
				t.Fatal(err)
			}

			bsonBytes, err := bson.Marshal(log)
			if err != nil {
				t.Fatal(err)
			}

			var parsedLog Log
			if err := bson.Unmarshal(bsonBytes, &parsedLog); err != nil {
				t.Fatalf("BSON Unmarshal failed with error: %v", err)
			}

			if parsedLog.Time.Before(time.Now().Add(-2 * time.Second)) {
				t.Error("The created_at time is not within the last 2 seconds")
			}
		})

		t.Run("test string method", func(t *testing.T) {
			logger := NewConsoleLogger(nil)
			str := log.String(logger)
			fmt.Printf("str: %v\n", str)

			now := time.Now().UTC()
			formattedTime := now.Format("2006-01-02 15:04:05")
			// NOTE: not sure if this will fail in some cases when running
			// remove the last 3 characters (seconds) from the formatted time
			formattedTime = formattedTime[:len(formattedTime)-3]
			if err := validate.StringIncludes(str, []string{
				"error",
				"my-source",
				"my-event",
				"qweqwe",
				formattedTime, // check if the printed time is the same as the current time,
			}); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("test omit empty fields", func(t *testing.T) {

		log := newLog(ERROR, "", "", "", "")

		decoded, err := utils.DecodeStructToStrings(log)
		if err != nil {
			t.Error(err)
		}

		shouldNotExistField := []string{
			"source",
			"event",
			"event_id",
		}

		for _, field := range shouldNotExistField {
			if strings.Contains(decoded.JSON, field) {
				t.Errorf("The field %v should not exist", field)
			}

			if strings.Contains(decoded.BSON, field) {
				t.Errorf("The field %v should not exist", field)
			}
		}
	})

	t.Run("test custom settings", func(t *testing.T) {

		// Format with no empty spaces used so that it's easier to parse it from
		// the log string
		const format = "20060102T150405Z"

		loc, err := time.LoadLocation("Europe/Berlin")
		if err != nil {
			t.Fatal(err)
		}

		logger := NewConsoleLogger(&LoggerSettings{
			Location:   loc,
			TimeFormat: format,
		}).SetSource("my-source")

		log := newLog(ERROR, "qweqwe", "my-source", "my-event", "")

		fmt.Printf("log: %v\n", log.String(logger))
		logString := log.String(logger)

		// get the string which is before the "error" string, and check if it is the same as the
		// formatted time
		parts := strings.SplitN(logString, "error", 2)
		if len(parts) < 2 {
			t.Fatalf("expected log string to have at least two parts, got: %v", logString)
		}
		timePart := parts[0]

		// replace all empty spaces with nothing
		timePart = strings.ReplaceAll(timePart, " ", "")

		fmt.Printf("timePart: %v\n", timePart)

		// parse the time from the timePart
		parsedDate, err := time.Parse(format, timePart)
		if err != nil {
			t.Fatal(err)
		}

		// check if the parsed date is the same as the current time
		if parsedDate.Before(time.Now().Add(-2 * time.Second)) {
			t.Fatal("The created_at time is not within the last 2 seconds")
		}
	})
}

func TestConsoleLogger(t *testing.T) {
	t.Run("test console logger", func(t *testing.T) {
		if NewConsoleLogger(nil).GetProps().Settings != nil {
			t.Error("Settings should be nil")
		}

		if NewConsoleLogger(nil).SetEvent("my-event").GetProps().Event != "my-event" {
			t.Error("SetEvent failed")
		}

		lg := NewConsoleLogger(nil).
			SetSource("my-source").
			SetEventID("my-event-id")

		if lg.GetProps().Source != "my-source" && lg.GetProps().EventID != "my-event-id" {
			t.Error("SetEventID failed")
		}

		logExists, err := NewConsoleLogger(nil).LogExists(nil)
		if err == nil {
			t.Error("LogExists should always return an error")
		}

		if logExists {
			t.Error("LogExists should always return false")
		}

		if err.Error() != "method cannot be used with ConsoleLogger" {
			t.Error("LogExists should always return a predefined error")
		}
	})
}

func TestMongoLogger(t *testing.T) {
	conn, err := mongodb.New("localhost", 27017, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Disconnect(context.Background())

	coll := mongodb.Coll(conn, "test", "test_mongo_logger")

	t.Run("test log creation", func(t *testing.T) {
		// Remove the previous data
		if err := coll.Drop(context.Background()); err != nil {
			t.Fatal(err)
		}

		logger := NewMongoLogger(coll, nil)
		if logger == nil {
			t.Error("NewMongoLogger should not return nil")
		}

		if err := logger.Debug("qwe"); err != nil {
			t.Error(err)
		}

		// find the log in the collection
		var log Log
		if err := coll.FindOne(context.Background(), bson.M{}).Decode(&log); err != nil {
			t.Error(err)
		}

		if log.Message != "qwe" {
			t.Error("The log message should be 'qwe'")
		}

		if log.Level != DEBUG {
			t.Error("The log level should be ", DEBUG)
		}

		if log.Source != "" {
			t.Error("The log source should be empty")
		}

		if log.Source != "" {
			t.Error("The log source should be empty")
		}

		if log.Event != "" {
			t.Error("The log event should be empty")
		}

		if log.EventID != "" {
			t.Error("The log event_id should be empty")
		}

		// if the time is not within the last 2 seconds
		if log.Time.Before(time.Now().Add(-2 * time.Second)) {
			t.Error("The created_at time is not within the last 2 seconds")
		}

		// decoded, err := utils.DecodeStructToStrings(log)
		// if err != nil {
		// 	t.Error(err)
		// }

		// fmt.Printf("decoded.JSON: %v\n", decoded.JSON)
		// fmt.Printf("decoded.BSON: %v\n", decoded.BSON)
	})

	t.Run("test log fields", func(t *testing.T) {
		coll := mongodb.Coll(conn, "test", "test_mongo_logger_with_fields")
		if err := coll.Drop(context.Background()); err != nil {
			t.Fatal(err)
		}

		logger := NewMongoLogger(coll, nil)

		var asd error

		if err := logger.Debug("qwe", Fields{
			"key1": "value1",
			"key2": 123,
			"asd":  asd,
		}); err != nil {
			t.Error(err)
		}

		var log Log
		if err := coll.FindOne(context.Background(), bson.M{}).Decode(&log); err != nil {
			t.Error(err)
		}

		fmt.Printf("log.Fields: %v\n", log.Fields)
		for k, v := range log.Fields {
			fmt.Printf("k: %-10v v: %-10v type: %-10T\n", k, v, v)
		}

		// test if the expected fields are in the log
		if log.Fields["key1"] != "value1" {
			t.Error("The key1 field should be 'value1', got: ", log.Fields["key1"])
		}

		// NOTE: i'm not sure what to do in this case, tests fail without the int32 type
		if log.Fields["key2"] != int32(123) {
			t.Error("The key2 field should be 123, got: ", log.Fields["key2"])
		}

		if log.Fields["asd"] != nil {
			t.Error("The asd field should be the same as the asd variable")
		}
	})

	t.Run("test log creation", func(t *testing.T) {
		// Remove the previous data
		if err := coll.Drop(context.Background()); err != nil {
			t.Fatal(err)
		}

		logger := NewMongoLogger(coll, nil).SetEventID("my-event-id")

		if err := logger.Info("my unique info event"); err != nil {
			t.Error(err)
		}

		t.Run("check if a created log exists", func(t *testing.T) {
			filter := bson.M{"event_id": "my-event-id"}
			exists, err := logger.LogExists(filter)
			if err != nil {
				t.Error(err)
			}

			if !exists {
				t.Error("The log should exist")
			}
		})

		t.Run("check if a non existent log does not exitst", func(t *testing.T) {
			filter := bson.M{"event_id": "this does not exist"}
			exists, err := logger.LogExists(filter)
			if err != nil {
				t.Error(err)
			}

			if exists {
				t.Error("The log should not exist")
			}
		})
	})

	t.Run("test find logs", func(t *testing.T) {
		// Remove the previous data
		if err := coll.Drop(context.Background()); err != nil {
			t.Fatal(err)
		}

		numLogs := 10

		logger := NewMongoLogger(coll, nil).SetEventID("my-event-id")
		for i := 0; i < numLogs; i++ {
			logger.Debug("this is a test")
		}

		// ---- test the find logs method ----
		test1, err := logger.FindLogs(LogFilter{
			Limit:   100,
			Skip:    0,
			EventID: "my-event-id",
		})

		if err != nil {
			t.Error(err)
		}

		if len(test1) != numLogs {
			t.Errorf("The number of logs should be %v", numLogs)
		}

		// if all of the logs are not debug level and the data is not "this is a test"
		// then the test failed
		for _, log := range test1 {
			if log.Level != DEBUG || log.Message != "this is a test" {
				t.Error("The logs are not correct")
			}
		}

		// ---- test the find logs method with a limit ----
		test2, err := logger.FindLogs(LogFilter{
			EventID: "my-event-id",
			Limit:   5,
			Skip:    0,
		})

		if err != nil {
			t.Error(err)
		}

		if len(test2) != 5 {
			t.Errorf("The number of logs should be %v", 5)
		}

		// ---- other filters ----
		test3, err := logger.FindLogs(LogFilter{
			EventID: "this-event-does-not-exist",
			Limit:   10,
		})

		if err != nil {
			t.Error(err)
		}

		if len(test3) != 0 {
			t.Errorf("The number of logs should be %v", 0)
		}
	})
}

func BenchmarkLogger(b *testing.B) {

	logger := NewConsoleLogger(nil).SetSource("qwe")
	_ = logger

	// b.Run("log creation", func(b *testing.B) {
	// 	logger.Debug("qwe", Fields{"asd": "asd"})
	// 	// for i := 0; i < b.N; i++ {
	// 	// }
	// })

	meta := Fields{"key1": "value1"}

	// // _ = new
	// _ = new.String(logger)

	for i := 0; i < b.N; i++ {
		new := newLog(ERROR, "qweqwe", "my-source", "my-event", "", meta)

		new.String(logger)

		// logger.Debug("qwe", Fields{"asd": "qwe"})
	}
}
