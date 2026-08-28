package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/overthinkinglabs/sqd/src/services/files"
)

func TestIgnoreListSkipsFilesByBaseName(t *testing.T) {
	ignoreList := files.NewIgnoreList()
	ignoreList.AddPattern("*.min.js")

	if !ignoreList.ShouldSkipFile("static/app.min.js", "app.min.js") {
		t.Error("expected *.min.js to skip app.min.js")
	}

	if ignoreList.ShouldSkipFile("static/app.js", "app.js") {
		t.Error("expected app.js not to be skipped")
	}
}

func TestIgnoreListSkipsDirectory(t *testing.T) {
	ignoreList := files.NewIgnoreList()
	ignoreList.AddPattern("vendor/")

	if !ignoreList.ShouldSkipDir("vendor") {
		t.Error("expected vendor/ to skip vendor directory")
	}

	if ignoreList.ShouldSkipDir("src") {
		t.Error("expected src directory not to be skipped")
	}
}

func TestIgnoreListRecursivePattern(t *testing.T) {
	ignoreList := files.NewIgnoreList()
	ignoreList.AddPattern("node_modules/**")

	if !ignoreList.ShouldSkipDir("node_modules") {
		t.Error("expected node_modules/** to skip node_modules directory")
	}

	if !ignoreList.ShouldSkipDir("node_modules/foo/bar") {
		t.Error("expected node_modules/** to skip nested node_modules directory")
	}

	if ignoreList.ShouldSkipDir("nested/node_modules") {
		t.Errorf("expected nested/node_modules not to be skipped by node_modules/**")
	}
}

func TestFinderRespectsIgnoreList(t *testing.T) {
	tmpDir := t.TempDir()

	filesToCreate := []string{"a.md", "b.min.js", "vendor/c.md", "node_modules/d.md"}
	for _, relativePath := range filesToCreate {
		fullPath := filepath.Join(tmpDir, relativePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0o644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	if err := os.WriteFile(filepath.Join(tmpDir, ".sqdignore"), []byte("*.min.js\nvendor/\nnode_modules/**\n"), 0o644); err != nil {
		t.Fatalf("failed to write sqdignore: %v", err)
	}

	ignoreList, err := files.LoadIgnoreList(".sqdignore")
	if err != nil {
		t.Fatalf("failed to load ignore list: %v", err)
	}

	finder := files.NewFinder(ignoreList)
	foundFiles, walkErr := finder.FindFiles("*.md")
	if walkErr != nil {
		t.Fatalf("unexpected walk error: %v", walkErr)
	}

	if len(foundFiles) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(foundFiles), foundFiles)
	}

	if len(foundFiles) == 0 || filepath.Base(foundFiles[0]) != "a.md" {
		t.Errorf("expected a.md, got %v", foundFiles)
	}
}
