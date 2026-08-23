package tests

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/commands"
	"github.com/overthinkinglabs/sqd/src/services/files"
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

	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)
	parallelizer := files.NewParallelizer(utils)
	sorter := commands.NewSorter()
	searcher := commands.NewSearcher(parallelizer, sorter, utils)
	counter := commands.NewCounter(parallelizer, searcher)

	return searcher, counter
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = writePipe
	fn()
	_ = writePipe.Close()
	os.Stdout = oldStdout

	data, _ := io.ReadAll(readPipe)
	_ = readPipe.Close()

	return string(data)
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

	stats, err := searcher.Select([]string{file}, command, models.TextOutput)
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
	stats, err := searcher.Select([]string{missingFile}, command, models.TextOutput)
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

	stats, err := searcher.Select([]string{file}, command, models.TextOutput)
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

	stats, err := searcher.Select([]string{file}, command, models.TextOutput)
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

func TestSearcherAppliesLimit(t *testing.T) {
	file := createSemanticTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT content FROM semantic_test.txt WHERE content LIKE '%a%' ORDER BY content LIMIT 2")

	output := captureOutput(t, func() {
		_, _ = searcher.Select([]string{file}, command, models.TextOutput)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 results, got %d", len(lines))
	}
}

func TestSearcherSelectLine(t *testing.T) {
	file := createSemanticTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT line FROM semantic_test.txt WHERE content = 'pattern line'")

	output := captureOutput(t, func() {
		_, _ = searcher.Select([]string{file}, command, models.TextOutput)
	})

	if strings.TrimSpace(output) != "2" {
		t.Errorf("expected line number '2', got %q", output)
	}
}

func TestSearcherRespectsSelectColumnOrder(t *testing.T) {
	file := createSemanticTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT content, line, name FROM semantic_test.txt WHERE content = 'pattern line'")

	output := captureOutput(t, func() {
		_, _ = searcher.Select([]string{file}, command, models.TextOutput)
	})

	ansiEscape := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanOutput := ansiEscape.ReplaceAllString(output, "")
	fields := strings.Split(strings.TrimSpace(cleanOutput), "\t")
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %q", len(fields), cleanOutput)
	}
	if fields[0] != "pattern line" {
		t.Errorf("expected content first, got %q", fields[0])
	}
	if fields[1] != "2" {
		t.Errorf("expected line second, got %q", fields[1])
	}
	if !strings.HasSuffix(fields[2], "semantic_test.txt") {
		t.Errorf("expected name last, got %q", fields[2])
	}
}

func TestSearcherOutputsJSON(t *testing.T) {
	file := createSemanticTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM semantic_test.txt WHERE content = 'pattern line'")

	output := captureOutput(t, func() {
		_, _ = searcher.Select([]string{file}, command, models.JSONOutput)
	})

	if !strings.Contains(output, `"name":`) {
		t.Errorf("expected JSON to contain name field, got %q", output)
	}

	if !strings.Contains(output, `"line":2`) {
		t.Errorf("expected JSON to contain line number 2, got %q", output)
	}

	if !strings.Contains(output, `"content":"pattern line"`) {
		t.Errorf("expected JSON to contain content, got %q", output)
	}
}

func TestSearcherOutputsCSV(t *testing.T) {
	file := createSemanticTestFile(t)
	searcher, _ := newSearcherAndCounter(t)

	parser := mock.NewParser()
	command := parser.Parse("SELECT line FROM semantic_test.txt WHERE content = 'pattern line'")

	output := captureOutput(t, func() {
		_, _ = searcher.Select([]string{file}, command, models.CSVOutput)
	})

	expected := "line\n2\n"
	if output != expected {
		t.Errorf("expected CSV %q, got %q", expected, output)
	}
}
