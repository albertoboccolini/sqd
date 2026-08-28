package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/overthinkinglabs/sqd/src/models"
)

type Utils struct {
	defaultConfig *models.DefaultConfig
}

func NewUtils(defaultConfig *models.DefaultConfig) *Utils {
	return &Utils{defaultConfig: defaultConfig}
}

func (utils *Utils) IsPathInsideCwd(path string) bool {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		return false
	}

	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	resolvedPath, _ := filepath.EvalSymlinks(absolutePath)
	if resolvedPath == "" {
		resolvedPath = absolutePath
	}

	relativePath, err := filepath.Rel(currentWorkingDir, resolvedPath)
	if err != nil {
		return false
	}

	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return false
	}

	return true
}

func (utils *Utils) CanWriteFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	defer func() { _ = file.Close() }()
	return true
}

func (utils *Utils) printLineCountMessage(count int, action string) {
	if count == 1 {
		fmt.Printf("1 line %s\n", action)
		return
	}

	fmt.Printf("%d lines %s\n", count, action)
}

func (utils *Utils) PrintUpdateMessage(count int) {
	utils.printLineCountMessage(count, "updated")
}

func (utils *Utils) PrintDeleteMessage(count int) {
	utils.printLineCountMessage(count, "deleted")
}

func (utils *Utils) PrintStats(stats models.ExecutionStats) {
	if !utils.defaultConfig.Output.ShowStats {
		return
	}

	elapsed := time.Since(stats.StartTime).Seconds()
	fmt.Printf("Processed: %d files in %.2fms\n", stats.Processed, elapsed*1000)
	if stats.Skipped > 0 {
		fmt.Printf("Skipped: %d files\n", stats.Skipped)
	}
}

func (utils *Utils) colorEscape() string {
	switch utils.defaultConfig.Output.Color {
	case "green":
		return "\033[1;32m"
	case "red":
		return "\033[1;31m"
	case "yellow":
		return "\033[1;33m"
	case "cyan":
		return "\033[1;36m"
	case "purple":
		return "\033[38;5;93m"
	case "none":
		return ""
	default:
		return "\033[1;34m"
	}
}

func (utils *Utils) HighlightMatch(text string, pattern *regexp.Regexp) string {
	if pattern == nil {
		return text
	}

	colorStart := utils.colorEscape()
	if colorStart == "" {
		return text
	}

	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		return colorStart + match + "\033[0m"
	})
}

func (utils *Utils) HighlightName(file string, pattern *regexp.Regexp) string {
	fileName := filepath.Base(file)
	baseDir := filepath.Dir(file)
	highlightedName := utils.HighlightMatch(fileName, pattern)
	return fmt.Sprintf("%s/%s", baseDir, highlightedName)
}

func (utils *Utils) AddWalkWarnings(errorCollection *models.ErrorCollection, walkWarnings error) {
	if walkWarnings == nil {
		return
	}

	if walkErrorCollection, ok := errors.AsType[*models.ErrorCollection](walkWarnings); ok {
		for _, walkErr := range walkErrorCollection.Errors() {
			errorCollection.Add(walkErr)
		}

		return
	}

	errorCollection.Add(walkWarnings)
}
