package services

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileFinder struct {
	maxTextFileSize int64
	bufferSize      int
}

func NewFileFinder() *FileFinder {
	return &FileFinder{
		maxTextFileSize: 100 * 1024 * 1024,
		bufferSize:      8000,
	}
}

// If the file cannot be stat'ed or opened, the function returns true so that
// callers like FindFiles do not silently skip those paths.
func (fileFinder *FileFinder) IsTextFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	if info.Size() > fileFinder.maxTextFileSize {
		return false
	}

	file, err := os.Open(path)
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

	var (
		files     []string
		mutex     sync.Mutex
		waitGroup sync.WaitGroup
		sem       = make(chan struct{}, 100)
	)

	filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if !matched {
			return nil
		}

		waitGroup.Add(1)
		sem <- struct{}{}

		go func(p string) {
			defer waitGroup.Done()
			defer func() { <-sem }()

			if fileFinder.IsTextFile(p) {
				mutex.Lock()
				files = append(files, p)
				mutex.Unlock()
			}
		}(path)

		return nil
	})

	waitGroup.Wait()
	return files
}
