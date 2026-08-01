package displayable_errors

import "fmt"

type FileProcessingError struct {
	*DisplayableError
}

func NewFileProcessingError(filename string, err error) *FileProcessingError {
	return &FileProcessingError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("Failed to process file %s: %v", filename, err),
		},
	}
}

func (fileProcessingError *FileProcessingError) Unwrap() error {
	return fileProcessingError.DisplayableError
}
