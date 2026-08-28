package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
)

func TestIsPathInsideCwdRelative(t *testing.T) {
	cwd, _ := os.Getwd()
	file := filepath.Join(cwd, "test.txt")
	os.WriteFile(file, []byte("test"), 0o644)
	defer os.Remove(file)

	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)

	if !utils.IsPathInsideCwd("./test.txt") {
		t.Error("relative path should be valid")
	}

	if !utils.IsPathInsideCwd("test.txt") {
		t.Error("relative path without ./ should be valid")
	}
}

func TestIsPathInsideCwdAbsolute(t *testing.T) {
	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)

	if utils.IsPathInsideCwd("/etc/passwd") {
		t.Error("absolute path outside cwd should be invalid")
	}
}

func TestIsPathInsideCwdTraversal(t *testing.T) {
	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)

	if utils.IsPathInsideCwd("../../../etc/passwd") {
		t.Error("path traversal should be blocked")
	}

	if utils.IsPathInsideCwd("..") {
		t.Error("parent directory should be blocked")
	}
}

func TestIsPathInsideCwdSymlink(t *testing.T) {
	cwd, _ := os.Getwd()
	symlink := filepath.Join(cwd, "test_symlink")
	os.Symlink("/tmp", symlink)
	defer os.Remove(symlink)

	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)

	if utils.IsPathInsideCwd(symlink) {
		t.Error("symlink outside cwd should be invalid")
	}
}

func TestHighlightMatchUsesConfigColor(t *testing.T) {
	colorConfig := models.DefaultConfig{
		Output: models.OutputConfig{Color: "green", ShowStats: true},
	}

	utils := services.NewUtils(&colorConfig)
	pattern := regexp.MustCompile("match")
	result := utils.HighlightMatch("this is a match", pattern)

	expected := "this is a \033[1;32mmatch\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHighlightMatchReturnsPlainTextWhenColorIsNone(t *testing.T) {
	colorConfig := models.DefaultConfig{
		Output: models.OutputConfig{Color: "none", ShowStats: true},
	}

	utils := services.NewUtils(&colorConfig)
	pattern := regexp.MustCompile("match")
	result := utils.HighlightMatch("this is a match", pattern)

	if result != "this is a match" {
		t.Errorf("expected plain text, got %q", result)
	}
}

func TestHighlightMatchFallsBackToBlueForUnknownColor(t *testing.T) {
	colorConfig := models.DefaultConfig{
		Output: models.OutputConfig{Color: "magenta", ShowStats: true},
	}

	utils := services.NewUtils(&colorConfig)
	pattern := regexp.MustCompile("match")
	result := utils.HighlightMatch("this is a match", pattern)

	expected := "this is a \033[1;34mmatch\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
