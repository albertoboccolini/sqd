package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/albertoboccolini/sqd/models"
)

const SQD_VERSION = "0.0.5"

type Writer interface {
	Printf(format string, args ...interface{})
	Fprintf(w io.Writer, format string, args ...interface{})
}

type StandardWriter struct{}

func (standardWriter *StandardWriter) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (standardWriter *StandardWriter) Fprintf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
}

type FileSystemOperations interface {
	Getwd() (string, error)
	Abs(path string) (string, error)
	EvalSymlinks(path string) (string, error)
	Rel(basepath, targpath string) (string, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

type OSFileSystemOperations struct{}

func (osFileSystemOperations *OSFileSystemOperations) Getwd() (string, error) {
	return os.Getwd()
}

func (osFileSystemOperations *OSFileSystemOperations) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func (osFileSystemOperations *OSFileSystemOperations) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

func (osFileSystemOperations *OSFileSystemOperations) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

func (osFileSystemOperations *OSFileSystemOperations) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

type Utils struct {
	writer     Writer
	filesystem FileSystemOperations
}

func NewUtils() *Utils {
	return &Utils{
		writer:     &StandardWriter{},
		filesystem: &OSFileSystemOperations{},
	}
}

func (utils *Utils) PrintUpdateMessage(total int) {
	utils.writer.Printf("Updated: %d occurrences\n", total)
}

func (utils *Utils) PrintProcessingErrorMessage(file string, err error) {
	utils.writer.Fprintf(os.Stderr, "%s: %v\n", file, err)
}

func (utils *Utils) PrintStats(stats models.ExecutionStats) {
	elapsed := time.Since(stats.StartTime).Seconds()
	utils.writer.Printf("Processed: %d files in %.2fms\n", stats.Processed, elapsed*1000)
	if stats.Skipped > 0 {
		utils.writer.Printf("Skipped: %d files\n", stats.Skipped)
	}
}

func (utils *Utils) IsPathInsideCwd(path string) bool {
	currentWorkingDir, err := utils.filesystem.Getwd()
	if err != nil {
		return false
	}

	absolutePath, err := utils.filesystem.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	resolvedPath, _ := utils.filesystem.EvalSymlinks(absolutePath)
	if resolvedPath == "" {
		resolvedPath = absolutePath
	}

	relativePath, err := utils.filesystem.Rel(currentWorkingDir, resolvedPath)
	if err != nil {
		return false
	}

	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return false
	}

	return true
}

func (utils *Utils) CanWriteFile(path string) bool {
	file, err := utils.filesystem.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	file.Close()
	return true
}
