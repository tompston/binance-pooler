package errgroup

import "strings"

// ErrGroup is a helper struct for cases when a single function
// could have multiple ErrGroup.
type ErrGroup []error

func New() *ErrGroup { return &ErrGroup{} }

func (eg *ErrGroup) Add(err error) {
	if eg != nil && err != nil {
		*eg = append(*eg, err)
	}
}

// Error implements the error interface. It returns a concatenated string of all
// non-nil ErrGroup, each separated by a semicolon.
func (eg *ErrGroup) Error() string {
	if eg == nil {
		return ""
	}

	var errStrings []string
	for _, err := range *eg {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	return strings.Join(errStrings, "; ")
}

func (eg *ErrGroup) Len() int {
	if eg == nil {
		return 0
	}

	return len(*eg)
}

// Return the error only if at least one of them happened. This is done because
// the ErrGroup is not nil when created, but it may be empty.
func (eg *ErrGroup) ToErr() error {
	if eg == nil {
		return nil
	}

	if eg.Len() == 0 {
		return nil
	}

	return eg
}
