package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/albertoboccolini/sqd/services"
)

func TestTransactionPreservesFilePermissions(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	os.WriteFile(file, []byte("content"), 0600)
	defer os.Remove(file)

	originalInfo, _ := os.Stat(file)
	originalMode := originalInfo.Mode()

	utils := services.NewUtils()
	sqlParser := services.NewSQLParser()

	command := sqlParser.Parse("UPDATE test.txt SET content='NEW' WHERE content = 'content'")

	fileOperator := services.NewFileOperator(utils)
	fileOperator.ExecuteCommand(command, []string{file}, true, false)

	newInfo, _ := os.Stat(file)
	if newInfo.Mode() != originalMode {
		t.Errorf("file permissions changed: %v -> %v", originalMode, newInfo.Mode())
	}
}

func TestTransactionEmptyFileHandling(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	os.WriteFile(file, []byte(""), 0644)
	defer os.Remove(file)

	utils := services.NewUtils()
	sqlParser := services.NewSQLParser()

	command := sqlParser.Parse("UPDATE test.txt SET content='NEW' WHERE content = 'nonexistent'")

	fileOperator := services.NewFileOperator(utils)
	fileOperator.ExecuteCommand(command, []string{file}, true, false)

	result, _ := os.ReadFile(file)
	if string(result) != "" {
		t.Error("empty file should remain empty when no matches")
	}
}

func TestTransactionWithTrailingNewline(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	content := "line1\nline2\nline3\n"
	os.WriteFile(file, []byte(content), 0644)
	defer os.Remove(file)

	utils := services.NewUtils()
	sqlParser := services.NewSQLParser()

	command := sqlParser.Parse("UPDATE test.txt SET content='UPDATED' WHERE content = 'line2'")

	fileOperator := services.NewFileOperator(utils)
	fileOperator.ExecuteCommand(command, []string{file}, true, false)

	result, _ := os.ReadFile(file)
	expected := "line1\nUPDATED\nline3\n"
	if string(result) != expected {
		t.Errorf("transaction should preserve file format: got %q, want %q", string(result), expected)
	}
}

func TestTransactionMultipleFilesSuccess(t *testing.T) {
	cwd, _ := os.Getwd()
	file1 := filepath.Join(cwd, "multi1.txt")
	file2 := filepath.Join(cwd, "multi2.txt")
	file3 := filepath.Join(cwd, "multi3.txt")

	os.WriteFile(file1, []byte("test"), 0644)
	os.WriteFile(file2, []byte("test"), 0644)
	os.WriteFile(file3, []byte("test"), 0644)

	defer os.Remove(file1)
	defer os.Remove(file2)
	defer os.Remove(file3)

	utils := services.NewUtils()
	sqlParser := services.NewSQLParser()

	command := sqlParser.Parse("UPDATE *.txt SET content='CHANGED' WHERE content LIKE 'test'")

	fileOperator := services.NewFileOperator(utils)
	fileOperator.ExecuteCommand(command, []string{file1, file2, file3}, true, false)

	result1, _ := os.ReadFile(file1)
	result2, _ := os.ReadFile(file2)
	result3, _ := os.ReadFile(file3)

	if string(result1) != "CHANGED" || string(result2) != "CHANGED" || string(result3) != "CHANGED" {
		t.Error("all files should be updated atomically in transaction")
	}

	for _, f := range []string{file1, file2, file3} {
		if _, err := os.Stat(f + ".sqd_backup"); err == nil {
			t.Errorf("backup for %s should be cleaned up after successful transaction", f)
		}
	}
}
