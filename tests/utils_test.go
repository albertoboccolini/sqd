package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/albertoboccolini/sqd/services"
)

func TestIsPathInsideCwdRelative(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	os.WriteFile(file, []byte("test"), 0644)
	defer os.Remove(file)

	if !services.IsPathInsideCwd("./test.txt") {
		t.Error("relative path should be valid")
	}

	if !services.IsPathInsideCwd("test.txt") {
		t.Error("relative path without ./ should be valid")
	}
}

func TestIsPathInsideCwdAbsolute(t *testing.T) {
	if services.IsPathInsideCwd("/etc/passwd") {
		t.Error("absolute path outside cwd should be invalid")
	}
}

func TestIsPathInsideCwdTraversal(t *testing.T) {
	if services.IsPathInsideCwd("../../../etc/passwd") {
		t.Error("path traversal should be blocked")
	}

	if services.IsPathInsideCwd("..") {
		t.Error("parent directory should be blocked")
	}
}

func TestIsPathInsideCwdSymlink(t *testing.T) {
	cwd, _ := os.Getwd()
	symlink := filepath.Join(cwd, "test_symlink")
	os.Symlink("/tmp", symlink)
	defer os.Remove(symlink)

	if services.IsPathInsideCwd(symlink) {
		t.Error("symlink outside cwd should be invalid")
	}
}
