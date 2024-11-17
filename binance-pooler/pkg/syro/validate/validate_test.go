package validate

import (
	"testing"
)

func TestEmptyStringsInStructExist(t *testing.T) {

	type SubStruct struct {
		Field1 string
		Field2 string
	}

	type MainStruct struct {
		SubStruct1 SubStruct
		SubStruct2 SubStruct
	}

	// Test case 1: No empty strings, should pass
	mainStruct1 := MainStruct{
		SubStruct1: SubStruct{
			Field1: "non-empty",
			Field2: "also non-empty",
		},
		SubStruct2: SubStruct{
			Field1: "another non-empty",
			Field2: "not empty",
		},
	}
	if err := EmptyStringsInStructExist(mainStruct1); err != nil {
		t.Errorf("Test case 1 failed: No empty strings expected, but error found: %s", err)
	}

	// Test case 2: Empty string in SubStruct2.Field2, should fail
	mainStruct2 := MainStruct{
		SubStruct1: SubStruct{
			Field1: "non-empty",
			Field2: "also non-empty",
		},
		SubStruct2: SubStruct{
			Field1: "another non-empty",
			Field2: "", // This will be considered empty
		},
	}
	expectedErrorMessage := "empty string found in SubStruct2.Field2"
	err := EmptyStringsInStructExist(mainStruct2)
	if err == nil {
		t.Error("Test case 2 failed: Expected error, but no error found.")
	}

	// Test case 3: Input is not a struct, should fail
	input := "not a struct"
	expectedErrorMessage = "expected a struct, got string"
	err = EmptyStringsInStructExist(input)
	if err == nil {
		t.Error("Test case 3 failed: Expected error, but no error found.")
	} else if err.Error() != expectedErrorMessage {
		t.Errorf("Test case 3 failed: Expected error message '%s', but got '%s'", expectedErrorMessage, err.Error())
	}
}

func TestStringIncludes(t *testing.T) {
	tests := []struct {
		input   string
		arr     []string
		wantErr bool
		errMsg  string
	}{
		{
			input:   "hello world",
			arr:     []string{"hello", "world"},
			wantErr: false,
		},
		{
			input:   "hello world",
			arr:     []string{"world", "goodbye"},
			wantErr: true,
			errMsg:  "input string 'hello world' does not include 'goodbye'",
		},
		{
			input:   "foobar",
			arr:     []string{"foo"},
			wantErr: false,
		},
		{
			input:   "foobar",
			arr:     []string{"bar", "baz"},
			wantErr: true,
			errMsg:  "input string 'foobar' does not include 'baz'",
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		err := StringIncludes(tt.input, tt.arr)
		if tt.wantErr {
			if err == nil {
				t.Errorf("StringIncludes(%q, %v) wanted error but got none", tt.input, tt.arr)
			} else if err.Error() != tt.errMsg {
				t.Errorf("StringIncludes(%q, %v) got error %q, want %q", tt.input, tt.arr, err, tt.errMsg)
			}
		} else if err != nil {
			t.Errorf("StringIncludes(%q, %v) got unexpected error %q", tt.input, tt.arr, err)
		}
	}
}

func TestArraysMatch(t *testing.T) {
	tests := []struct {
		expected, actual []string
		wantErr          bool
		errMessage       string
	}{
		{
			expected:   []string{"a", "b", "c"},
			actual:     []string{"a", "b", "c"},
			wantErr:    false,
			errMessage: "",
		},
		{
			expected:   []string{"a", "b", "c"},
			actual:     []string{"a", "x", "c"},
			wantErr:    true,
			errMessage: "array did not have the expected column name at position 1",
		},
		{
			expected:   []string{"a", "b"},
			actual:     []string{"a", "b", "c"},
			wantErr:    true,
			errMessage: "array did not have the expected number of columns, 3",
		},
	}

	for _, test := range tests {
		err := ArraysMatch(test.expected, test.actual)
		if test.wantErr {
			if err == nil {
				t.Errorf("expected an error for input %v and %v, got none", test.expected, test.actual)
			} else if err.Error() != test.errMessage {
				t.Errorf("expected error message: %q, got: %q", test.errMessage, err.Error())
			}
		} else if err != nil {
			t.Errorf("did not expect an error for input %v and %v, got: %v", test.expected, test.actual, err)
		}
	}
}
