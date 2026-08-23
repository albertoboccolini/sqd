package files

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type ignorePattern struct {
	pattern     string
	isRecursive bool
	matchesPath bool
}

type IgnoreList struct {
	patterns []ignorePattern
}

func NewIgnoreList() *IgnoreList {
	return &IgnoreList{patterns: make([]ignorePattern, 0)}
}

func LoadIgnoreList(path string) (*IgnoreList, error) {
	ignoreList := NewIgnoreList()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoreList, nil
		}

		return nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		ignoreList.AddPattern(line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreList, nil
}

func (ignoreList *IgnoreList) AddPattern(pattern string) {
	cleaned := strings.TrimSpace(pattern)
	if cleaned == "" {
		return
	}

	isRecursive := strings.Contains(cleaned, "**")
	matchesPath := strings.Contains(cleaned, "/")

	ignoreList.patterns = append(ignoreList.patterns, ignorePattern{
		pattern:     cleaned,
		isRecursive: isRecursive,
		matchesPath: matchesPath,
	})
}

func (ignoreList *IgnoreList) IsEmpty() bool {
	return len(ignoreList.patterns) == 0
}

func (ignoreList *IgnoreList) ShouldSkipDir(relativePath string) bool {
	pathParts := strings.Split(relativePath, string(filepath.Separator))

	for _, pattern := range ignoreList.patterns {
		if ignoreList.patternMatchesDir(pattern, pathParts) {
			return true
		}
	}

	return false
}

func (ignoreList *IgnoreList) ShouldSkipFile(relativePath string, baseName string) bool {
	for _, pattern := range ignoreList.patterns {
		if ignoreList.patternMatchesFile(pattern, relativePath, baseName) {
			return true
		}
	}

	return false
}

func (ignoreList *IgnoreList) patternMatchesDir(pattern ignorePattern, pathParts []string) bool {
	effectivePattern := strings.TrimSuffix(pattern.pattern, "/")
	patternParts := strings.Split(effectivePattern, "/")

	for length := 1; length <= len(pathParts); length++ {
		if matchGlob(patternParts, pathParts[:length]) {
			return true
		}
	}

	return false
}

func (ignoreList *IgnoreList) patternMatchesFile(pattern ignorePattern, relativePath string, baseName string) bool {
	if strings.HasSuffix(pattern.pattern, "/") {
		return false
	}

	if !pattern.matchesPath {
		matched, _ := filepath.Match(pattern.pattern, baseName)
		return matched
	}

	patternParts := strings.Split(pattern.pattern, "/")
	pathParts := strings.Split(relativePath, string(filepath.Separator))
	return matchGlob(patternParts, pathParts)
}

func matchGlob(patternParts, pathParts []string) bool {
	for len(patternParts) > 0 {
		patternPart := patternParts[0]
		patternParts = patternParts[1:]

		if patternPart == "**" {
			for consumed := 0; consumed <= len(pathParts); consumed++ {
				if matchGlob(patternParts, pathParts[consumed:]) {
					return true
				}
			}

			return false
		}

		if len(pathParts) == 0 {
			return false
		}

		matched, err := filepath.Match(patternPart, pathParts[0])
		if err != nil || !matched {
			return false
		}

		pathParts = pathParts[1:]
	}

	return len(pathParts) == 0
}
