package commands

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

type lineHandler func(line string, lineNumber int) error

func readFileLines(file string, handler lineHandler) error {
	fileHandle, err := os.Open(file)
	if err != nil {
		return err
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
			return readErr
		}
	}
}
