package displayable_errors

type InvalidQueryError struct {
	*DisplayableError
}

func NewInvalidQueryError(description string) *InvalidQueryError {
	return &InvalidQueryError{
		DisplayableError: &DisplayableError{
			Description: description,
		},
	}
}

func (invalidQueryError *InvalidQueryError) Unwrap() error {
	return invalidQueryError.DisplayableError
}
