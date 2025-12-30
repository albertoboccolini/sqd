package services

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Open(name string) (*os.File, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type OSFileSystem struct{}

func (osFileSystem *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (osFileSystem *OSFileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (osFileSystem *OSFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

type FileFinder struct {
	filesystem      FileSystem
	maxTextFileSize int64
	bufferSize      int
}

func NewFileFinder() *FileFinder {
	return &FileFinder{
		filesystem:      &OSFileSystem{},
		maxTextFileSize: 100 * 1024 * 1024,
		bufferSize:      8000,
	}
}

// If the file cannot be stat'ed or opened, the function returns true so that
// callers like FindFiles do not silently skip those paths.
func (fileFinder *FileFinder) IsTextFile(path string) bool {
	info, err := fileFinder.filesystem.Stat(path)
	if err != nil {
		return true
	}

	if info.Size() > fileFinder.maxTextFileSize {
		return false
	}

	file, err := fileFinder.filesystem.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, fileFinder.bufferSize)
	n, _ := file.Read(buf)

	for _, b := range buf[:n] {
		if b == 0 {
			return false
		}

		if b < 9 {
			return false
		}
	}

	return true
}

func (fileFinder *FileFinder) FindFiles(pattern string) []string {
	if !strings.Contains(pattern, "*") {
		return []string{pattern}
	}

	var files []string

	fileFinder.filesystem.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched && fileFinder.IsTextFile(path) {
			files = append(files, path)
		}

		return nil
	})

	return files
}
