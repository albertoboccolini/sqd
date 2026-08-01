package displayable_errors

import "fmt"

type PermissionDeniedError struct {
	*DisplayableError
}

func NewPermissionDeniedError(filename string) *PermissionDeniedError {
	return &PermissionDeniedError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("Permission denied: %s", filename),
		},
	}
}

func (permissionDeniedError *PermissionDeniedError) Unwrap() error {
	return permissionDeniedError.DisplayableError
}
