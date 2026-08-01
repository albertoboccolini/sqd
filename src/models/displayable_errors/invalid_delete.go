package displayable_errors

type InvalidDeleteError struct {
	*DisplayableError
}

func NewInvalidDeleteError(description string) *InvalidDeleteError {
	return &InvalidDeleteError{
		DisplayableError: &DisplayableError{
			Description: description,
		},
	}
}

func (invalidDeleteError *InvalidDeleteError) Unwrap() error {
	return invalidDeleteError.DisplayableError
}
