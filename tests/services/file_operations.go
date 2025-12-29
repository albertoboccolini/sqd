package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/albertoboccolini/sqd/services"
)

func TestTransactionUpdateSuccess(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	content := "line1\nline2\nline3"
	os.WriteFile(file, []byte(content), 0644)
	defer os.Remove(file)

	cmd := services.ParseSQL("UPDATE test.txt SET content='UPDATED' WHERE content = 'line2'")
	services.ExecuteCommand(cmd, []string{file}, true)

	result, _ := os.ReadFile(file)
	if string(result) != "line1\nUPDATED\nline3" {
		t.Errorf("transaction update failed: got %s", string(result))
	}

	backupPath := file + ".sqd_backup"
	if _, err := os.Stat(backupPath); err == nil {
		t.Error("backup file should be cleaned up after successful transaction")
	}
}

func TestTransactionDeleteSuccess(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	content := "keep1\nremove\nkeep2"
	os.WriteFile(file, []byte(content), 0644)
	defer os.Remove(file)

	cmd := services.ParseSQL("DELETE FROM test.txt WHERE content = 'remove'")
	services.ExecuteCommand(cmd, []string{file}, true)

	result, _ := os.ReadFile(file)
	if string(result) != "keep1\nkeep2" {
		t.Errorf("transaction delete failed: got %s", string(result))
	}
}

func TestTransactionRollbackOnWriteError(t *testing.T) {
	cwd, _ := os.Getwd()
	file1 := filepath.Join(cwd, "test1.txt")
	file2 := filepath.Join(cwd, "test2.txt")

	originalContent1 := "original1"
	originalContent2 := "original2"

	os.WriteFile(file1, []byte(originalContent1), 0644)
	os.WriteFile(file2, []byte(originalContent2), 0444)
	defer os.Remove(file1)
	defer os.Remove(file2)

	cmd := services.ParseSQL("UPDATE *.txt SET content='NEW' WHERE content LIKE '%original%'")
	services.ExecuteCommand(cmd, []string{file1, file2}, true)

	result1, _ := os.ReadFile(file1)
	if string(result1) != originalContent1 {
		t.Error("file1 should be rolled back to original content")
	}

	if _, err := os.Stat(file1 + ".sqd_backup"); err == nil {
		t.Error("backup should be cleaned up after rollback")
	}
}

func TestTransactionBatchUpdateSuccess(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	content := "foo\nbar\nbaz"
	os.WriteFile(file, []byte(content), 0644)
	defer os.Remove(file)

	sql := `UPDATE test.txt SET content="FOO" WHERE content = "foo", SET content="BAR" WHERE content = "bar"`
	cmd := services.ParseSQL(sql)
	services.ExecuteCommand(cmd, []string{file}, true)

	result, _ := os.ReadFile(file)
	if string(result) != "FOO\nBAR\nbaz" {
		t.Errorf("batch transaction failed: got %s", string(result))
	}
}

func TestTransactionBatchDeleteSuccess(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	content := "keep\ndelete1\nkeep2\ndelete2"
	os.WriteFile(file, []byte(content), 0644)
	defer os.Remove(file)

	sql := `DELETE FROM test.txt WHERE content = "delete1", WHERE content = "delete2"`
	cmd := services.ParseSQL(sql)
	services.ExecuteCommand(cmd, []string{file}, true)

	result, _ := os.ReadFile(file)
	if string(result) != "keep\nkeep2" {
		t.Errorf("batch delete transaction failed: got %s", string(result))
	}
}

func TestTransactionMultipleFilesAtomic(t *testing.T) {
	cwd, _ := os.Getwd()
	file1 := filepath.Join(cwd, "atomic1.txt")
	file2 := filepath.Join(cwd, "atomic2.txt")
	file3 := filepath.Join(cwd, "atomic3.txt")

	os.WriteFile(file1, []byte("test1"), 0644)
	os.WriteFile(file2, []byte("test2"), 0644)
	os.WriteFile(file3, []byte("test3"), 0444)

	defer os.Remove(file1)
	defer os.Remove(file2)
	defer os.Remove(file3)

	cmd := services.ParseSQL("UPDATE *.txt SET content='CHANGED' WHERE content LIKE '%test%'")
	services.ExecuteCommand(cmd, []string{file1, file2, file3}, true)

	result1, _ := os.ReadFile(file1)
	result2, _ := os.ReadFile(file2)

	if string(result1) != "test1" || string(result2) != "test2" {
		t.Error("all files should be rolled back when one fails")
	}
}

func TestTransactionPreservesFilePermissions(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	os.WriteFile(file, []byte("content"), 0600)
	defer os.Remove(file)

	originalInfo, _ := os.Stat(file)
	originalMode := originalInfo.Mode()

	cmd := services.ParseSQL("UPDATE test.txt SET content='NEW' WHERE content = 'content'")
	services.ExecuteCommand(cmd, []string{file}, true)

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

	cmd := services.ParseSQL("UPDATE test.txt SET content='NEW' WHERE content = 'nonexistent'")
	services.ExecuteCommand(cmd, []string{file}, true)

	result, _ := os.ReadFile(file)
	if string(result) != "" {
		t.Error("empty file should remain empty when no matches")
	}
}
