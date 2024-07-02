// Package convert holds bunch of util functions that can be used
// for converting input
package convert

import (
	"reflect"
)

// StructStringFieldsToStrings converts a struct to a slice of strings. Only string fields are converted.
func StructStringFieldsToStrings(s any) []string {
	v := reflect.ValueOf(s)

	var strings []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String {
			strings = append(strings, field.String())
		}
	}

	return strings
}
