package validate

import (
	"fmt"
	"reflect"
	"strings"
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

// StringIncludes checks if the input string contains all of the strings in
// the array. If the input string does not contain a strings from the
// array, an error is returned.
func StringIncludes(input string, arr []string) error {
	for _, str := range arr {
		if !strings.Contains(input, str) {
			return fmt.Errorf("input string '%s' does not include '%s'", input, str)
		}
	}
	return nil
}
