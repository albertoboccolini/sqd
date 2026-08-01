package displayable_errors

import "fmt"

type FileReadError struct {
	*DisplayableError
}

func NewFileReadError(filename string, err error) *FileReadError {
	return &FileReadError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("Unable to open file %s: %v", filename, err),
		},
	}
}

func (fileReadError *FileReadError) Unwrap() error {
	return fileReadError.DisplayableError
}
