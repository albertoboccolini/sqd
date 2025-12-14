package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PrintUpdateMessage(total int) {
	fmt.Printf("Updated: %d occurrences\n", total)
}

func isPathInsideCwd(path string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	eval, _ := filepath.EvalSymlinks(abs)
	if eval == "" {
		eval = abs
	}

	rel, err := filepath.Rel(cwd, eval)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}
