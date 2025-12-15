package services

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const maxTextFileSize = 100 * 1024 * 1024

// IsTextFile checks if a file is a text file by reading its content and looking for non-text bytes.
// This is necessary because some files may have text-like extensions but contain binary data.
func IsTextFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.Size() > maxTextFileSize {
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		return false
	}

	defer file.Close()
	buf := make([]byte, 8000)
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

func FindFiles(pattern string) []string {
	if !strings.Contains(pattern, "*") {
		if !isPathInsideCwd(pattern) {
			return []string{}
		}
		return []string{pattern}
	}

	var files []string

	filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		if !isPathInsideCwd(path) {
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
