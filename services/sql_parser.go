package services

import (
	"regexp"
	"strings"

	"sqd/models"
)

func ParseSQL(sql string) models.Command {
	sql = strings.TrimSpace(sql)
	upper := strings.ToUpper(sql)

	var cmd models.Command

	// Detection action (già esistente)
	if strings.HasPrefix(upper, "SELECT COUNT") {
		cmd.Action = "COUNT"
	}
	if strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "SELECT COUNT") {
		cmd.Action = "SELECT"
	}
	if strings.HasPrefix(upper, "UPDATE") {
		cmd.Action = "UPDATE"
	}
	if strings.HasPrefix(upper, "DELETE") {
		cmd.Action = "DELETE"
	}

	if cmd.Action == "UPDATE" {
		cmd.File = extractBetween(sql, "UPDATE", "SET")
	}

	if cmd.Action == "DELETE" {
		cmd.File = extractBetween(sql, "DELETE FROM", "WHERE")
	}

	if cmd.Action == "SELECT" || cmd.Action == "COUNT" {
		cmd.File = extractBetween(sql, "FROM", "WHERE")
	}

	cmd.File = strings.TrimSpace(cmd.File)

	if strings.Contains(upper, "WHERE CONTENT =") {
		cmd.MatchExact = true
		exactMatch := extractAfter(sql, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")
		cmd.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
	}

	if strings.Contains(upper, "WHERE CONTENT LIKE") {
		cmd.MatchExact = false
		likePattern := extractAfter(sql, "LIKE")
		likePattern = strings.Trim(likePattern, " '\"")
		cmd.Pattern = likeToRegex(likePattern)
	}

	if cmd.Action == "UPDATE" {
		cmd.Replace = extractBetween(sql, "SET content=", "WHERE")
		cmd.Replace = strings.Trim(cmd.Replace, "'\"")
	}

	return cmd
}

func extractBetween(s, start, end string) string {
	startUpper := strings.ToUpper(start)
	endUpper := strings.ToUpper(end)
	sUpper := strings.ToUpper(s)

	startIdx := strings.Index(sUpper, startUpper)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startUpper)

	endIdx := strings.Index(sUpper[startIdx:], endUpper)
	if endIdx == -1 {
		return strings.TrimSpace(s[startIdx:])
	}

	return strings.TrimSpace(s[startIdx : startIdx+endIdx])
}

func extractAfter(s, marker string) string {
	markerUpper := strings.ToUpper(marker)
	sUpper := strings.ToUpper(s)

	idx := strings.Index(sUpper, markerUpper)
	if idx == -1 {
		return ""
	}

	return strings.TrimSpace(s[idx+len(markerUpper):])
}

func likeToRegex(pattern string) *regexp.Regexp {
	hasStart := strings.HasPrefix(pattern, "%")
	hasEnd := strings.HasSuffix(pattern, "%")

	pattern = strings.Trim(pattern, "%")
	pattern = regexp.QuoteMeta(pattern)

	if !hasStart && hasEnd {
		pattern = "^" + pattern
	}
	if hasStart && !hasEnd {
		pattern = pattern + "$"
	}

	return regexp.MustCompile(pattern)
}
