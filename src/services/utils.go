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

type Utils struct{}

func NewUtils() *Utils {
	return &Utils{}
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

func (utils *Utils) PrintUpdateMessage(count int) {
	if count == 1 {
		fmt.Println("1 line updated")
		return
	}

	fmt.Printf("%d lines updated\n", count)
}

func (utils *Utils) PrintDeleteMessage(count int) {
	if count == 1 {
		fmt.Println("1 line deleted")
		return
	}

	fmt.Printf("%d lines deleted\n", count)
}

func (utils *Utils) PrintStats(stats models.ExecutionStats) {
	elapsed := time.Since(stats.StartTime).Seconds()
	fmt.Printf("Processed: %d files in %.2fms\n", stats.Processed, elapsed*1000)
	if stats.Skipped > 0 {
		fmt.Printf("Skipped: %d files\n", stats.Skipped)
	}
}

func (utils *Utils) HighlightMatch(text string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		return "\033[1;34m" + match + "\033[0m"
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

	var walkErrorCollection *models.ErrorCollection
	if errors.As(walkWarnings, &walkErrorCollection) {
		for _, walkErr := range walkErrorCollection.Errors() {
			errorCollection.Add(walkErr)
		}

		return
	}

	errorCollection.Add(walkWarnings)
}
