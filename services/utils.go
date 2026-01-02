package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/albertoboccolini/sqd/models"
)

const SQD_VERSION = "0.0.7"

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

func (utils *Utils) printUpdateMessage(total int) {
	fmt.Printf("Updated: %d occurrences\n", total)
}

func (utils *Utils) printProcessingErrorMessage(file string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
}

func (utils *Utils) printStats(stats models.ExecutionStats) {
	elapsed := time.Since(stats.StartTime).Seconds()
	fmt.Printf("Processed: %d files in %.2fms\n", stats.Processed, elapsed*1000)
	if stats.Skipped > 0 {
		fmt.Printf("Skipped: %d files\n", stats.Skipped)
	}
}

func (utils *Utils) canWriteFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	file.Close()
	return true
}
