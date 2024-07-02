package errgroup

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	eg := New()
	if eg == nil {
		t.Errorf("New() should not return nil")
	}
	if len(*eg) != 0 {
		t.Errorf("New() should initialize an empty ErrGroup")
	}
}

func TestAdd(t *testing.T) {
	eg := New()
	err1 := errors.New("first error")
	err2 := errors.New("second error")

	eg.Add(err1)
	if len(*eg) != 1 {
		t.Errorf("Add() did not properly add the first error")
	}

	eg.Add(err2)
	if len(*eg) != 2 {
		t.Errorf("Add() did not properly add the second error")
	}

	eg.Add(nil) // test adding nil error
	if len(*eg) != 2 {
		t.Errorf("Add() should not add nil errors")
	}
}

func TestError(t *testing.T) {
	eg := New()
	err1 := errors.New("first error")
	err2 := errors.New("second error")

	eg.Add(err1)
	eg.Add(err2)

	expected := "first error; second error"
	if eg.Error() != expected {
		t.Errorf("Error() returned %q, want %q", eg.Error(), expected)
	}

	eg = New() // test with no errors
	if eg.Error() != "" {
		t.Errorf("Error() should return an empty string for an empty ErrGroup, got %q", eg.Error())
	}
}

func TestLen(t *testing.T) {
	eg := New()
	if eg.Len() != 0 {
		t.Errorf("Len() should return 0 for a new ErrGroup, got %d", eg.Len())
	}

	eg.Add(errors.New("first error"))
	if eg.Len() != 1 {
		t.Errorf("Len() should return 1 after adding one error, got %d", eg.Len())
	}

	eg.Add(nil) // adding nil should not change the count
	if eg.Len() != 1 {
		t.Errorf("Len() should still return 1 after adding nil, got %d", eg.Len())
	}
}
