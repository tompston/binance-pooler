package utils

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// PrettyPrint formats a decoded json or bson string
func PrettyPrint(data string) (string, error) {
	var i any

	// Try to unmarshal as JSON first
	if json.Unmarshal([]byte(data), &i) != nil {
		// If JSON fails, try BSON
		if bson.Unmarshal([]byte(data), &i) != nil {
			return "", fmt.Errorf("failed to unmarshal data as JSON or BSON")
		}
	}

	pretty, err := json.MarshalIndent(i, "", "  ")
	return string(pretty), err
}

type DecodedStrings struct {
	JSON string
	BSON string
}

// Run the struct through json and bson marshalers and return the results as strings.
func DecodeStructToStrings(v any) (*DecodedStrings, error) {
	json, err := json.Marshal(&v)
	if err != nil {
		return nil, err
	}

	bson, err := bson.MarshalExtJSON(&v, false, false)
	if err != nil {
		return nil, err
	}

	return &DecodedStrings{JSON: string(json), BSON: string(bson)}, nil
}

func LogIfArgExists(msg any, loggerFn []func(any)) {
	if len(loggerFn) == 1 {
		loggerFn[0](msg)
	}
}
