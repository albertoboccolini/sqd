package tests

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/tests/mock"
)

func TestParallelProcessingWithErrors(t *testing.T) {
	cwd, _ := os.Getwd()

	file1 := filepath.Join(cwd, "parallel_valid1.txt")
	file2 := filepath.Join(cwd, "parallel_valid2.txt")
	invalidFile := filepath.Join(cwd, "nonexistent_file.txt")

	os.WriteFile(file1, []byte("content1\ncontent1\n"), 0o644)
	os.WriteFile(file2, []byte("content2\ncontent2\ncontent2\n"), 0o644)
	defer os.Remove(file1)
	defer os.Remove(file2)

	files := []string{file1, invalidFile, file2}

	dispatcher := mock.NewDispatcher()
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT * FROM parallel_*.txt WHERE content LIKE 'content'")

	err := dispatcher.Execute(command, files, false, false, false)
	if err == nil {
		t.Fatal("expected an error for missing file")
	}

	var errorCollection *models.ErrorCollection
	if !errors.As(err, &errorCollection) {
		t.Fatalf("expected ErrorCollection, got %T", err)
	}
	if len(errorCollection.Errors()) != 1 {
		t.Errorf("expected 1 error, got %d", len(errorCollection.Errors()))
	}

	if _, err := os.Stat(file1); err != nil {
		t.Error("valid file1 should still exist after processing")
	}

	if _, err := os.Stat(file2); err != nil {
		t.Error("valid file2 should still exist after processing")
	}

	if _, err := os.Stat(invalidFile); err == nil {
		t.Error("invalid file should not exist")
	}
}

func TestParallelProcessingConcurrencyLimit(t *testing.T) {
	cwd, _ := os.Getwd()

	numFiles := models.MAX_CONCURRENT_GOROUTINES * 2
	files := make([]string, numFiles)

	for i := range numFiles {
		file := filepath.Join(cwd, "concurrent_test_"+string(rune(i%26+'a'))+string(rune(i/26+'0'))+".txt")
		os.WriteFile(file, []byte("test line\n"), 0o644)
		files[i] = file
		defer os.Remove(file)
	}

	dispatcher := mock.NewDispatcher()
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT * FROM concurrent_test_*.txt WHERE content LIKE 'test'")

	if err := dispatcher.Execute(command, files, false, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, file := range files {
		if _, err := os.Stat(file); err != nil {
			t.Errorf("file %d should exist after high concurrency processing: %v", i, err)
		}
	}
}

func TestParallelProcessingMaintainsFileIntegrity(t *testing.T) {
	cwd, _ := os.Getwd()

	numFiles := models.MAX_CONCURRENT_GOROUTINES * 2
	files := make([]string, numFiles)
	expectedContent := make(map[string]string)

	for i := range numFiles {
		file := filepath.Join(cwd, "integrity_test_"+string(rune('a'+i))+".txt")
		content := strings.Repeat("line "+string(rune('0'+i))+"\n", i+1)
		os.WriteFile(file, []byte(content), 0o644)
		files[i] = file
		expectedContent[file] = content
		defer os.Remove(file)
	}

	dispatcher := mock.NewDispatcher()
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT * FROM integrity_test_*.txt WHERE content LIKE 'line'")

	if err := dispatcher.Execute(command, files, false, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for file, expected := range expectedContent {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("failed to read file %s: %v", file, err)
			continue
		}

		if string(data) != expected {
			t.Errorf("file %s content corrupted during parallel processing", file)
		}
	}
}
