package displayable_errors

type DisplayableError struct {
	Description string
}

func (displayableError *DisplayableError) Error() string {
	return displayableError.Description
}

func (displayableError *DisplayableError) Unwrap() error {
	return nil
}
