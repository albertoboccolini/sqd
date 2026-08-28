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

func (extractor *Extractor) compileExact(literal string) *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.QuoteMeta(literal) + "$")
}

func (extractor *Extractor) compileLike(pattern string) *regexp.Regexp {
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

	switch {
	case hasStart && hasEnd:
		return regexp.MustCompile(pattern)
	case hasStart:
		return regexp.MustCompile(pattern + "$")
	case hasEnd:
		return regexp.MustCompile("^" + pattern)
	default:
		return regexp.MustCompile(pattern)
	}
}
