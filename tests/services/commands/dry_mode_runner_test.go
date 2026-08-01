package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/tests/mock"
)

func TestValidateTransactionModeStopsOnFirstError(t *testing.T) {
	dryRunner := mock.NewDryModeRunner()

	pattern := regexp.MustCompile("test")
	command := models.Command{
		Action:  models.UPDATE,
		Pattern: pattern,
		Replace: "changed",
	}

	stats := &models.ExecutionStats{}
	cwd, _ := os.Getwd()
	validFile := filepath.Join(cwd, "valid.txt")
	invalidFile := filepath.Join(cwd, "..", "invalid.txt")

	err := dryRunner.Validate(command, []string{invalidFile, validFile}, stats, true, false)

	if err == nil {
		t.Error("expected transaction mode to return error on invalid path")
	}
}

func TestValidateNonTransactionModeContinuesAfterError(t *testing.T) {
	cwd, _ := os.Getwd()
	validFile := filepath.Join(cwd, "valid.txt")
	os.WriteFile(validFile, []byte("content\n"), 0o644)
	defer os.Remove(validFile)

	dryRunner := mock.NewDryModeRunner()
	pattern := regexp.MustCompile("content")
	command := models.Command{
		Action:  models.UPDATE,
		Pattern: pattern,
		Replace: "changed",
	}

	stats := &models.ExecutionStats{}
	invalidFile := filepath.Join(cwd, "..", "invalid.txt")

	err := dryRunner.Validate(command, []string{invalidFile, validFile}, stats, false, false)

	if err == nil {
		t.Error("expected error collection to be returned with invalid file error")
	}

	if stats.Processed != 1 {
		t.Errorf("expected 1 file processed, got %d", stats.Processed)
	}

	if stats.Skipped != 1 {
		t.Errorf("expected 1 file skipped, got %d", stats.Skipped)
	}
}

func TestValidatePermissionDenied(t *testing.T) {
	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "readonly.txt")
	os.WriteFile(testFile, []byte("content\n"), 0o400)
	defer os.Remove(testFile)

	dryRunner := mock.NewDryModeRunner()

	pattern := regexp.MustCompile("content")
	command := models.Command{
		Action:  models.UPDATE,
		Pattern: pattern,
		Replace: "changed",
	}

	stats := &models.ExecutionStats{}

	err := dryRunner.Validate(command, []string{testFile}, stats, false, false)

	if err == nil {
		t.Error("expected error for permission denied")
	}

	if stats.Skipped != 1 {
		t.Errorf("expected 1 skipped for permission denied, got %d", stats.Skipped)
	}

	if stats.Processed != 0 {
		t.Errorf("expected 0 processed for permission denied, got %d", stats.Processed)
	}
}
