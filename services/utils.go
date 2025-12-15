package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const SQD_VERSION = "0.0.3"

func PrintUpdateMessage(total int) {
	fmt.Printf("Updated: %d occurrences\n", total)
}

func isPathInsideCwd(path string) bool {
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

func canWriteFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	file.Close()
	return true
}
