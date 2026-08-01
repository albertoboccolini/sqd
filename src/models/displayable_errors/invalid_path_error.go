package displayable_errors

import "fmt"

type InvalidPathError struct {
	*DisplayableError
}

func NewInvalidPathError(filename string) *InvalidPathError {
	return &InvalidPathError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("Invalid path: %s", filename),
		},
	}
}

func (invalidPathError *InvalidPathError) Unwrap() error {
	return invalidPathError.DisplayableError
}
