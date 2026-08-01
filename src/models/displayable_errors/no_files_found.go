package displayable_errors

import "fmt"

type NoFilesFoundError struct {
	*DisplayableError
}

func NewNoFilesFoundError(pattern string) *NoFilesFoundError {
	return &NoFilesFoundError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("No files matched the pattern: %s", pattern),
		},
	}
}

func (noFilesFoundError *NoFilesFoundError) Unwrap() error {
	return noFilesFoundError.DisplayableError
}
