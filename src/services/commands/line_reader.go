package commands

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
)

type LineReader struct{}

func NewLineReader() *LineReader {
	return &LineReader{}
}

func (lineReader *LineReader) MatchesCondition(line string, pattern *regexp.Regexp, negate bool) bool {
	if pattern == nil {
		return true
	}

	matches := pattern.MatchString(line)
	if negate {
		matches = !matches
	}

	return matches
}

func (lineReader *LineReader) ReadFileLines(file string, handler func(line string, lineNumber int) error) error {
	fileHandle, err := os.Open(file)
	if err != nil {
		return lineReader.MapReadError(file, err)
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

func (lineReader *LineReader) MapReadError(file string, err error) error {
	if errors.Is(err, os.ErrPermission) {
		return displayable_errors.NewPermissionDeniedError(file)
	}

	return displayable_errors.NewFileReadError(file, err)
}

func (lineReader *LineReader) MergeErrors(errorCollection *models.ErrorCollection, source *models.ErrorCollection) {
	for _, err := range source.Errors() {
		errorCollection.Add(err)
	}
}

func (lineReader *LineReader) ErrorCollectionOrNil(errorCollection *models.ErrorCollection) error {
	if errorCollection.HasErrors() {
		return errorCollection
	}

	return nil
}
