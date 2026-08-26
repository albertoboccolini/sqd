package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

func setupPermissionDeniedTest(t *testing.T) ([]string, error) {
	tmpDir := t.TempDir()

	fileA := filepath.Join(tmpDir, "file_a.md")
	if err := os.WriteFile(fileA, []byte("content A"), 0o644); err != nil {
		t.Fatalf("failed to create file_a.md: %v", err)
	}

	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	fileBPath := filepath.Join(subdir, "file_b.md")
	if err := os.WriteFile(fileBPath, []byte("content B"), 0o644); err != nil {
		t.Fatalf("failed to create file_b.md: %v", err)
	}

	if err := os.Chmod(subdir, 0o000); err != nil {
		t.Fatalf("failed to chmod subdir: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(subdir, 0o755); err != nil {
			t.Errorf("failed to restore subdir permissions: %v", err)
		}
		if err := os.Chdir(cwd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	ignoreList := files.NewIgnoreList()
	finder := files.NewFinder(ignoreList)
	return finder.FindFiles("*.md")
}

func TestIsTextFileText(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.WriteString("This is plain text\n")
	file.Close()

	ignoreList := files.NewIgnoreList()
	finder := files.NewFinder(ignoreList)

	if !finder.IsTextFile(file.Name()) {
		t.Error("text file should be detected as text")
	}
}

func TestIsTextFileBinary(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.bin")
	defer os.Remove(file.Name())
	file.Write([]byte{0x00, 0x01, 0xFF, 0xFE, 0x00, 0x00})
	file.Close()

	ignoreList := files.NewIgnoreList()
	finder := files.NewFinder(ignoreList)

	if finder.IsTextFile(file.Name()) {
		t.Error("binary file should not be detected as text")
	}
}

func TestIsTextFileNullByte(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.WriteString("text\x00more")
	file.Close()

	ignoreList := files.NewIgnoreList()
	finder := files.NewFinder(ignoreList)

	if finder.IsTextFile(file.Name()) {
		t.Error("file with null byte should not be text")
	}
}

func TestIsTextFileControlChars(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.Write([]byte{0x01, 0x02, 0x03})
	file.Close()

	ignoreList := files.NewIgnoreList()
	finder := files.NewFinder(ignoreList)

	if finder.IsTextFile(file.Name()) {
		t.Error("file with control chars should not be text")
	}
}

func TestFindFilesReturnsFilesFoundBeforePermissionError(t *testing.T) {
	foundFiles, _ := setupPermissionDeniedTest(t)

	if len(foundFiles) == 0 {
		t.Errorf("should find file_a.md before permission denied on subdir")
	}
}

func TestFindFilesReturnsErrorCollectionOnPermissionDenied(t *testing.T) {
	_, err := setupPermissionDeniedTest(t)

	if err == nil {
		t.Errorf("should return error with walk warnings")
		return
	}

	if _, ok := errors.AsType[*models.ErrorCollection](err); !ok {
		t.Errorf("should return error collection with walk errors")
	}
}

func TestFindFilesIncludesWalkErrorInCollection(t *testing.T) {
	_, err := setupPermissionDeniedTest(t)

	var errorCollection *models.ErrorCollection
	if !errors.As(err, &errorCollection) {
		t.Fatalf("error should be ErrorCollection")
	}

	if !errorCollection.HasErrors() {
		t.Errorf("error collection should have errors")
		return
	}

	found := false
	for _, e := range errorCollection.Errors() {
		if _, ok := errors.AsType[*displayable_errors.WalkError](e); ok {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("should contain walk error in error collection")
	}
}
