package services

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func IsTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8000)
	n, _ := f.Read(buf)

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

func FindFiles(pattern string) []string {
	if !strings.Contains(pattern, "*") {
		return []string{pattern}
	}

	var files []string

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched && IsTextFile(path) {
			files = append(files, path)
		}

		return nil
	})

	return files
}
