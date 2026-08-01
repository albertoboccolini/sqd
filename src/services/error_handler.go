package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
)

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (errorHandler *ErrorHandler) handleSingleError(err error) {
	var displayableError *displayable_errors.DisplayableError

	if errors.As(err, &displayableError) {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// TODO consider logging the error to a file for later analysis
	fmt.Fprintf(os.Stderr, "Fatal error: %v. If this persists, open an issue on GitHub.\n", err)
}

func (errorHandler *ErrorHandler) HandleError(err error) {
	var errorCollection *models.ErrorCollection

	if errors.As(err, &errorCollection) {
		for _, e := range errorCollection.Errors() {
			errorHandler.handleSingleError(e)
		}

		return
	}

	errorHandler.handleSingleError(err)
}
