package services

import (
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

func ParseSQL(sql string) models.Command {
	sql = strings.TrimSpace(sql)
	upperSql := strings.ToUpper(sql)

	command := parseAction(sql, upperSql)
	parseWhereClause(&command, sql, upperSql)
	parseSetClause(&command, sql, upperSql)

	return command
}

func parseAction(sql, upperSql string) models.Command {
	var command models.Command

	if strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.COUNT
		command.File = extractBetween(sql, "FROM", "WHERE")
		return command
	}

	if strings.HasPrefix(upperSql, "SELECT") {
		command.Action = models.SELECT
		command.Columns = extractColumns(sql)
		command.File = extractBetween(sql, "FROM", "WHERE")
		return command
	}

	if strings.HasPrefix(upperSql, "UPDATE") {
		command.Action = models.UPDATE
		command.File = extractBetween(sql, "UPDATE", "SET")
		return command
	}

	if strings.HasPrefix(upperSql, "DELETE") {
		command.Action = models.DELETE
		command.File = extractBetween(sql, "DELETE FROM", "WHERE")
	}

	return command
}

func parseWhereClause(command *models.Command, sql, upperSql string) {
	if !strings.Contains(upperSql, "WHERE") {
		return
	}

	command.File = strings.TrimSpace(command.File)

	if parseBatchOperations(command, sql, upperSql) {
		return
	}

	if parseWhereNameClause(command, sql, upperSql) {
		if command.Action == models.DELETE {
			command.OperateOnName = true
		}
		return
	}

	parseWhereContentClause(command, sql, upperSql)
}

func parseBatchOperations(command *models.Command, sql, upperSql string) bool {
	if command.Action == models.UPDATE && strings.Count(upperSql, "SET CONTENT=") > 1 {
		command.IsBatch = true
		command.Replacements = parseBatchReplacements(sql)
		return true
	}

	if command.Action == models.DELETE && strings.Count(upperSql, "WHERE CONTENT =") > 1 {
		command.IsBatch = true
		command.Deletions = parseBatchDeletions(sql)
		return true
	}

	return false
}

func parseWhereNameClause(command *models.Command, sql, upperSql string) bool {
	if strings.Contains(upperSql, "WHERE NAME =") || strings.Contains(upperSql, "WHERE NAME=") {
		command.FilterOnName = true
		command.MatchExact = true
		command.Pattern = extractExactNamePattern(sql, upperSql)
		return true
	}

	if strings.Contains(upperSql, "WHERE NAME LIKE") {
		command.FilterOnName = true
		command.MatchExact = false
		likePattern := extractAfter(sql, "LIKE")
		likePattern = strings.Trim(likePattern, " '\"")
		command.Pattern = likeToRegex(likePattern)
		return true
	}

	return false
}

func parseWhereContentClause(command *models.Command, sql, upperSql string) {
	if strings.Contains(upperSql, "WHERE CONTENT =") {
		command.MatchExact = true
		exactMatch := extractAfter(sql, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")
		command.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		return
	}

	if strings.Contains(upperSql, "WHERE CONTENT LIKE") {
		command.MatchExact = false
		likePattern := extractAfter(sql, "LIKE")
		likePattern = strings.Trim(likePattern, " '\"")
		command.Pattern = likeToRegex(likePattern)
	}
}

func parseSetClause(command *models.Command, sql, upperSql string) {
	if command.Action != models.UPDATE {
		return
	}

	if strings.Contains(upperSql, "SET NAME") {
		command.OperateOnName = true
		command.Replace = extractBetween(sql, "SET name=", "WHERE")
		command.Replace = strings.Trim(command.Replace, "'\"")
		return
	}

	if !command.IsBatch {
		command.Replace = extractBetween(sql, "SET content=", "WHERE")
		command.Replace = strings.Trim(command.Replace, "'\"")
	}
}

func extractExactNamePattern(sql, upperSql string) *regexp.Regexp {
	idx := strings.Index(upperSql, "WHERE NAME")
	afterWhere := sql[idx+len("WHERE NAME"):]
	afterWhere = strings.TrimSpace(afterWhere)

	if strings.HasPrefix(afterWhere, "=") {
		afterWhere = strings.TrimSpace(afterWhere[1:])
	}

	exactMatch := strings.Trim(afterWhere, " '\"")
	return regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
}

func parseBatchDeletions(sql string) []models.Deletion {
	var deletions []models.Deletion

	parts := strings.Split(sql, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		upper := strings.ToUpper(part)

		if !strings.Contains(upper, "WHERE CONTENT =") {
			continue
		}

		exactMatch := extractAfter(part, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")

		deletions = append(deletions, models.Deletion{
			Pattern:    regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$"),
			MatchExact: true,
		})
	}

	return deletions
}

func parseBatchReplacements(sql string) []models.Replacement {
	var replacements []models.Replacement

	parts := strings.Split(sql, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		upperPart := strings.ToUpper(part)

		if !strings.Contains(upperPart, "SET CONTENT=") {
			continue
		}

		repl := parseSingleReplacement(part, upperPart)
		replacements = append(replacements, repl)
	}

	return replacements
}

func parseSingleReplacement(part, upperPart string) models.Replacement {
	replaceValue := extractBetween(part, "SET content=", "WHERE")
	replaceValue = strings.Trim(replaceValue, " '\"")

	repl := models.Replacement{Replace: replaceValue}

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

	return repl
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

func extractColumns(sql string) []string {
	cols := extractBetween(sql, "SELECT", "FROM")
	cols = strings.TrimSpace(cols)

	if cols == "" || cols == "*" {
		return []string{"*"}
	}

	parts := strings.Split(cols, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

func likeToRegex(pattern string) *regexp.Regexp {
	hasStart := strings.HasPrefix(pattern, "%")
	hasEnd := strings.HasSuffix(pattern, "%")

	pattern = strings.Trim(pattern, "%")
	pattern = regexp.QuoteMeta(pattern)

	if !hasStart && !hasEnd {
		pattern = "^" + pattern + "$"
	}

	if !hasStart && hasEnd {
		pattern = "^" + pattern
	}

	if hasStart && !hasEnd {
		pattern = pattern + "$"
	}

	return regexp.MustCompile(pattern)
}
