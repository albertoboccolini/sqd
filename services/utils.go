package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

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

	workingDirInfo, err := os.Stat(currentWorkingDir)
	if err != nil {
		return false
	}

	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return false
	}

	workingDirDevice := workingDirInfo.Sys().(*syscall.Stat_t).Dev
	fileDevice := fileInfo.Sys().(*syscall.Stat_t).Dev

	return workingDirDevice == fileDevice
}

func canWriteFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	file.Close()
	return true
}
