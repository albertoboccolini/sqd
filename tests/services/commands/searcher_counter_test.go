package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/services"
	"github.com/overthinkinglabs/sqd/services/commands"
	"github.com/overthinkinglabs/sqd/services/files"
	"github.com/overthinkinglabs/sqd/tests/mock"
)

func createLineReaderTestFile(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	file := filepath.Join(cwd, "line_reader_test.txt")
	if err := os.WriteFile(file, []byte("one\npattern line\nthree\n"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(file)
	})

	return file
}

func newSearcherAndCounter(t *testing.T) (*commands.Searcher, *commands.Counter) {
	t.Helper()

	utils := services.NewUtils()
	parallelizer := files.NewParallelizer(utils)
	sorter := commands.NewSorter()
	searcher := commands.NewSearcher(parallelizer, sorter, utils)
	counter := commands.NewCounter(parallelizer, searcher)

	return searcher, counter
}

func TestCounterCountsMatchingContentWithoutLoadingFileIntoMemory(t *testing.T) {
	file := createLineReaderTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM line_reader_test.txt WHERE content LIKE '%pattern%'")

	count, stats, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed file, got %d", stats.Processed)
	}
	if stats.Skipped != 0 {
		t.Errorf("expected 0 skipped files, got %d", stats.Skipped)
	}
}

func TestCounterCountsAllLinesWithoutLoadingFileIntoMemory(t *testing.T) {
	file := createLineReaderTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM line_reader_test.txt WHERE content LIKE '%'")

	count, stats, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed file, got %d", stats.Processed)
	}
}

func TestSearcherSelectsMatchingContentWithoutLoadingFileIntoMemory(t *testing.T) {
	file := createLineReaderTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT content FROM line_reader_test.txt WHERE content LIKE '%pattern%'")

	stats, err := searcher.Select([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed file, got %d", stats.Processed)
	}
	if stats.Skipped != 0 {
		t.Errorf("expected 0 skipped files, got %d", stats.Skipped)
	}
}

func TestSearcherReportsMissingFiles(t *testing.T) {
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT content FROM line_reader_test.txt WHERE content LIKE '%pattern%'")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	missingFile := filepath.Join(cwd, "line_reader_missing.txt")
	stats, err := searcher.Select([]string{missingFile}, command)
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
	if stats.Processed != 0 {
		t.Errorf("expected 0 processed files, got %d", stats.Processed)
	}
	if stats.Skipped != 1 {
		t.Errorf("expected 1 skipped file, got %d", stats.Skipped)
	}
}

func TestSearcherSelectsAllContentWhenFilteringByName(t *testing.T) {
	file := createLineReaderTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM line_reader_test.txt WHERE name LIKE 'line_reader%'")

	stats, err := searcher.Select([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed file, got %d", stats.Processed)
	}
	if stats.Skipped != 0 {
		t.Errorf("expected 0 skipped files, got %d", stats.Skipped)
	}
}

func TestCounterCountsWithAndClause(t *testing.T) {
	file := createLineReaderTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM line_reader_test.txt WHERE content LIKE '%pattern%' AND content LIKE '%line%'")

	count, stats, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed file, got %d", stats.Processed)
	}
}

func TestSearcherSelectsWithAndClause(t *testing.T) {
	file := createLineReaderTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT content FROM line_reader_test.txt WHERE content LIKE '%pattern%' AND content LIKE '%line%'")

	stats, err := searcher.Select([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed file, got %d", stats.Processed)
	}
	if stats.Skipped != 0 {
		t.Errorf("expected 0 skipped files, got %d", stats.Skipped)
	}
}

func createSemanticTestFile(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	file := filepath.Join(cwd, "semantic_test.txt")
	if err := os.WriteFile(file, []byte("pat\npattern line\none\nthree\nmodern\n"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(file)
	})

	return file
}

func TestCounterCountsExactContentMatch(t *testing.T) {
	file := createSemanticTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()

	command := parser.Parse("SELECT COUNT(*) FROM semantic_test.txt WHERE content = 'pattern line'")
	count, _, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 for exact match, got %d", count)
	}

	command = parser.Parse("SELECT COUNT(*) FROM semantic_test.txt WHERE content = 'pattern'")
	count, _, err = counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0 for exact match with no matching line, got %d", count)
	}
}

func TestCounterCountsPrefixLikeMatch(t *testing.T) {
	file := createSemanticTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM semantic_test.txt WHERE content LIKE 'patt%'")

	count, _, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 for prefix LIKE, got %d", count)
	}
}

func TestCounterCountsSuffixLikeMatch(t *testing.T) {
	file := createSemanticTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM semantic_test.txt WHERE content LIKE '%odern'")

	count, _, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 for suffix LIKE, got %d", count)
	}
}

func TestCounterCountsNegatedExactContentMatch(t *testing.T) {
	file := createSemanticTestFile(t)
	_, counter := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM semantic_test.txt WHERE content != 'pat'")

	count, _, err := counter.Count([]string{file}, command)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 4 {
		t.Errorf("expected count 4 for negated exact match, got %d", count)
	}
}
