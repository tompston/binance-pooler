package validate

import (
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// EmptyStringsInStructExist checks if there are empty strings in a struct.
func EmptyStringsInStructExist(v any) error {
	isEmptyString := func(val string) bool {
		return len(strings.TrimSpace(val)) == 0
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, got %T", v)
	}

	// Helper function to traverse nested structs and find empty strings
	var checkFields func(reflect.Value, string) error

	checkFields = func(v reflect.Value, path string) error {
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := v.Type().Field(i)
			fieldPath := path + "." + fieldType.Name

			switch fieldValue.Kind() {
			case reflect.String:
				if isEmptyString(fieldValue.String()) {
					return fmt.Errorf("empty string found in %s", fieldPath)
				}

			case reflect.Struct:
				if err := checkFields(fieldValue, fieldPath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return checkFields(value, "")
}

func ValueIsInArray[T comparable](i T, arr []T) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}

// StringsMatch compares two strings and returns an error if they are not equal.
func StringsMatch(s1, s2 string) error {
	if !reflect.DeepEqual(s1, s2) {
		for i := 0; i < len(s1) && i < len(s2); i++ {
			if s1[i] != s2[i] {
				return fmt.Errorf("difference at index %d: Value 1 has '%c', Value 2 has '%c'", i, s1[i], s2[i])
			}
		}
	}
	return nil
}

// StringIncludes checks if the input string contains all of the strings in the array.
// If the input string does not contain a strings from the array, an error is returned.
func StringIncludes(input string, arr []string) error {
	for _, str := range arr {
		if !strings.Contains(input, str) {
			return fmt.Errorf("input string '%s' does not include '%s'", input, str)
		}
	}
	return nil
}

func BsonMEqual(a, b bson.M) bool {
	if len(a) != len(b) {
		return false
	}

	for k := range a {
		if !reflect.DeepEqual(a[k], b[k]) {
			return false
		}
	}
	return true
}

// BsonSlicesEqual checks if two slices of bson.M are equal
func BsonSlicesEqual(a, b []bson.M) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !BsonMEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func ArraysMatch(expected, actual []string, verbose ...bool) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("array did not have the expected number of columns, %v", len(actual))
	}

	shouldReturnVerboseError := false
	if len(verbose) == 1 {
		shouldReturnVerboseError = verbose[0]
	}

	for i, exp := range expected {
		if actual[i] != exp {
			if shouldReturnVerboseError {
				return fmt.Errorf("array did not have the expected column name at position %v, encountered: %v, expected: %v", i, actual[i], exp)
			}
			return fmt.Errorf("array did not have the expected column name at position %v", i)
		}
	}

	return nil
}
