package sql

import (
	"regexp"
	"strings"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (extractor *Extractor) extractFilename(sql string, startKeyword, endKeyword string) string {
	upperSql := strings.ToUpper(sql)
	startIdx := strings.Index(upperSql, startKeyword)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(startKeyword)

	endIdx := strings.Index(upperSql[startIdx:], endKeyword)
	var filename string

	if endIdx == -1 {
		filename = strings.TrimSpace(sql[startIdx:])
	} else {
		filename = strings.TrimSpace(sql[startIdx : startIdx+endIdx])
	}

	return strings.Trim(filename, "'\"")
}

func (extractor *Extractor) likeToRegex(pattern string) *regexp.Regexp {
	hasStart := strings.HasPrefix(pattern, "%")
	hasEnd := strings.HasSuffix(pattern, "%")

	if hasStart {
		pattern = pattern[1:]
	}

	if hasEnd && len(pattern) > 0 {
		pattern = pattern[:len(pattern)-1]
	}

	if pattern == "" {
		return regexp.MustCompile(".*")
	}

	pattern = regexp.QuoteMeta(pattern)

	if hasStart && hasEnd {
		return regexp.MustCompile(pattern)
	}

	if hasStart {
		return regexp.MustCompile(pattern + "$")
	}

	if hasEnd {
		return regexp.MustCompile("^" + pattern)
	}

	return regexp.MustCompile(pattern)
}

func (extractor *Extractor) LikeSubstrings(pattern string) []string {
	parts := strings.Split(pattern, "%")
	substrings := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			substrings = append(substrings, part)
		}
	}

	return substrings
}
