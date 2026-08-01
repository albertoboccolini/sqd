package tests

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

func TestErrorHandler_DisplayableError(t *testing.T) {
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer pipeReader.Close()
	os.Stderr = pipeWriter

	err = displayable_errors.NewFileReadError("file.txt", errors.New("file not found"))

	handler := services.NewErrorHandler()
	handler.HandleError(err)

	pipeWriter.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(pipeReader); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	output := buffer.String()

	expected := "Unable to open file file.txt: file not found\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestErrorHandler_GenericError(t *testing.T) {
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer pipeReader.Close()
	os.Stderr = pipeWriter

	err = errors.New("generic error")

	handler := services.NewErrorHandler()
	handler.HandleError(err)

	pipeWriter.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(pipeReader); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	output := buffer.String()

	expected := "Fatal error: generic error. If this persists, open an issue on GitHub.\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
