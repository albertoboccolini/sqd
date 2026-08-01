package displayable_errors

import "fmt"

type TransactionFailedError struct {
	*DisplayableError
}

func NewTransactionFailedError(details string) *TransactionFailedError {
	return &TransactionFailedError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("Transaction failed: %s", details),
		},
	}
}

func (transactionFailedError *TransactionFailedError) Unwrap() error {
	return transactionFailedError.DisplayableError
}
