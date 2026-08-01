package commands

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/models/displayable_errors"
)

type lineHandler func(line string, lineNumber int) error

func readFileLines(file string, handler lineHandler) error {
	fileHandle, err := os.Open(file)
	if err != nil {
		return mapReadError(file, err)
	}
	defer func() { _ = fileHandle.Close() }()

	reader := bufio.NewReaderSize(fileHandle, 1024*1024)
	lineNumber := 0
	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			lineNumber++
			line = strings.TrimSuffix(line, "\n")
			handlerErr := handler(line, lineNumber)
			if handlerErr != nil {
				if errors.Is(handlerErr, io.EOF) {
					return nil
				}

				return handlerErr
			}
		}

		if readErr == io.EOF {
			return nil
		}

		if readErr != nil {
			return displayable_errors.NewFileReadError(file, readErr)
		}
	}
}

func mapReadError(file string, err error) error {
	if errors.Is(err, os.ErrPermission) {
		return displayable_errors.NewPermissionDeniedError(file)
	}

	return displayable_errors.NewFileReadError(file, err)
}

func mergeErrors(errorCollection *models.ErrorCollection, source *models.ErrorCollection) {
	for _, err := range source.Errors() {
		errorCollection.Add(err)
	}
}

func errorCollectionOrNil(errorCollection *models.ErrorCollection) error {
	if errorCollection.HasErrors() {
		return errorCollection
	}

	return nil
}
