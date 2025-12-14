package services

import (
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

func ParseSQL(sql string) models.Command {
	sql = strings.TrimSpace(sql)
	upperSql := strings.ToUpper(sql)

	var command models.Command

	if strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.COUNT
		command.File = extractBetween(sql, "FROM", "WHERE")
	}

	if strings.HasPrefix(upperSql, "SELECT") && !strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.SELECT
		command.File = extractBetween(sql, "FROM", "WHERE")
	}

	if strings.HasPrefix(upperSql, "UPDATE") {
		command.Action = models.UPDATE
		command.File = extractBetween(sql, "UPDATE", "SET")
	}

	if strings.HasPrefix(upperSql, "DELETE") {
		command.Action = models.DELETE
		command.File = extractBetween(sql, "DELETE FROM", "WHERE")
	}

	command.File = strings.TrimSpace(command.File)

	if command.Action == models.UPDATE && strings.Count(upperSql, "SET CONTENT=") > 1 {
		command.IsBatch = true
		command.Replacements = parseBatchReplacements(sql)
		return command
	}

	if command.Action == models.DELETE && strings.Count(upperSql, "WHERE CONTENT =") > 1 {
		command.IsBatch = true
		command.Deletions = parseBatchDeletions(sql)
		return command
	}

	if strings.Contains(upperSql, "WHERE CONTENT =") {
		command.MatchExact = true
		exactMatch := extractAfter(sql, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")
		command.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
	}

	if strings.Contains(upperSql, "WHERE CONTENT LIKE") {
		command.MatchExact = false
		likePattern := extractAfter(sql, "LIKE")
		likePattern = strings.Trim(likePattern, " '\"")
		command.Pattern = likeToRegex(likePattern)
	}

	if command.Action == models.UPDATE {
		command.Replace = extractBetween(sql, "SET content=", "WHERE")
		command.Replace = strings.Trim(command.Replace, "'\"")
	}

	return command
}

func parseBatchDeletions(sql string) []models.Deletion {
	var deletions []models.Deletion

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upper := strings.ToUpper(part)

		if !strings.Contains(upper, "WHERE CONTENT =") {
			continue
		}

		var del models.Deletion
		del.MatchExact = true

		exactMatch := extractAfter(part, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")
		del.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")

		deletions = append(deletions, del)
	}

	return deletions
}

func parseBatchReplacements(sql string) []models.Replacement {
	var replacements []models.Replacement

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upperPart := strings.ToUpper(part)

		if !strings.Contains(upperPart, "SET CONTENT=") {
			continue
		}

		var repl models.Replacement

		replaceValue := extractBetween(part, "SET content=", "WHERE")
		replaceValue = strings.Trim(replaceValue, " '\"")
		repl.Replace = replaceValue

		if strings.Contains(upperPart, "WHERE CONTENT =") {
			repl.MatchExact = true
			exactMatch := extractAfter(part, "WHERE content =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			repl.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if strings.Contains(upperPart, "WHERE CONTENT LIKE") {
			repl.MatchExact = false
			likePattern := extractAfter(part, "LIKE")
			likePattern = strings.Trim(likePattern, " '\"")
			repl.Pattern = likeToRegex(likePattern)
		}

		replacements = append(replacements, repl)
	}

	return replacements
}

func extractBetween(query, start, end string) string {
	upperStart := strings.ToUpper(start)
	upperEnd := strings.ToUpper(end)
	upperQuery := strings.ToUpper(query)

	startIndex := strings.Index(upperQuery, upperStart)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(upperStart)
	endIndex := strings.Index(upperQuery[startIndex:], upperEnd)

	if endIndex == -1 {
		return strings.TrimSpace(query[startIndex:])
	}

	return strings.TrimSpace(query[startIndex : startIndex+endIndex])
}

func extractAfter(query, marker string) string {
	markerUpper := strings.ToUpper(marker)
	upperQuery := strings.ToUpper(query)

	index := strings.Index(upperQuery, markerUpper)
	if index == -1 {
		return ""
	}

	return strings.TrimSpace(query[index+len(markerUpper):])
}

func likeToRegex(pattern string) *regexp.Regexp {
	if len(pattern) > 1000 {
		pattern = pattern[:1000]
	}

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
