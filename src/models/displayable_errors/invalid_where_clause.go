package displayable_errors

type InvalidWhereClauseError struct {
	*DisplayableError
}

func NewInvalidWhereClauseError(description string) *InvalidWhereClauseError {
	return &InvalidWhereClauseError{
		DisplayableError: &DisplayableError{
			Description: description,
		},
	}
}

func (invalidWhereClauseError *InvalidWhereClauseError) Unwrap() error {
	return invalidWhereClauseError.DisplayableError
}
