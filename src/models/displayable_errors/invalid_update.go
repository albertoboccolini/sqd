package displayable_errors

type InvalidUpdateError struct {
	*DisplayableError
}

func NewInvalidUpdateError(description string) *InvalidUpdateError {
	return &InvalidUpdateError{
		DisplayableError: &DisplayableError{
			Description: description,
		},
	}
}

func (invalidUpdateError *InvalidUpdateError) Unwrap() error {
	return invalidUpdateError.DisplayableError
}
