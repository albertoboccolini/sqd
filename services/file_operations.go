package services

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

func isFileBlocked(filename string) bool {
	if !isPathInsideCwd(filename) {
		return true
	}

	if !canWriteFile(filename) {
		return true
	}

	return false
}

func ExecuteCommand(command models.Command, files []string) {
	if command.Action == models.COUNT {
		total := 0
		for _, file := range files {
			total += countMatches(file, command.Pattern)
		}

		fmt.Printf("%d lines matched\n", total)
		return
	}

	if command.Action == models.SELECT {
		for _, file := range files {
			selectMatches(file, command.Pattern)
		}

		return
	}

	if command.Action == models.UPDATE {
		total := 0
		if command.IsBatch {
			for _, file := range files {
				total += updateFileInBatch(file, command.Replacements)
			}

			PrintUpdateMessage(total)
			return
		}

		for _, file := range files {
			total += updateFile(file, command.Pattern, command.Replace)
		}

		PrintUpdateMessage(total)
		return
	}

	if command.Action == models.DELETE {
		total := 0

		if command.IsBatch {
			for _, file := range files {
				total += deleteMatchesInBatch(file, command.Deletions)
			}

			fmt.Printf("Deleted: %d lines\n", total)
			return
		}

		for _, file := range files {
			total += deleteMatches(file, command.Pattern)
		}

		fmt.Printf("Deleted: %d lines\n", total)
	}
}

func countMatches(filename string, pattern *regexp.Regexp) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for _, line := range lines {
		if pattern.MatchString(line) {
			count++
		}
	}

	return count
}

func selectMatches(filename string, pattern *regexp.Regexp) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if pattern.MatchString(line) {
			fmt.Printf("%s:%d: %s\n", filename, i+1, line)
		}
	}
}

func updateFile(filename string, pattern *regexp.Regexp, replace string) int {
	if isFileBlocked(filename) {
		return 0
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for i, line := range lines {
		if pattern.MatchString(line) {
			lines[i] = pattern.ReplaceAllLiteralString(line, replace)
			count++
		}
	}

	if count > 0 {
		os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
	}

	return count
}

// updateFileInBatch applies multiple replacements to the file in a single pass.
// This is more efficient than applying each replacement separately.
func updateFileInBatch(filename string, replacements []models.Replacement) int {
	if isFileBlocked(filename) {
		return 0
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for i, line := range lines {
		for _, replacement := range replacements {
			if replacement.Pattern.MatchString(line) {
				lines[i] = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
				count++
				break
			}
		}
	}

	if count > 0 {
		os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
	}

	return count
}

func deleteMatches(filename string, pattern *regexp.Regexp) int {
	if isFileBlocked(filename) {
		return 0
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	filtered := []string{}
	count := 0

	for _, line := range lines {
		if !pattern.MatchString(line) {
			filtered = append(filtered, line)
			continue
		}
		count++
	}

	if count > 0 {
		os.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
	}

	return count
}

// deleteMatchesInBatch applies multiple deletions to the file in a single pass.
// This is more efficient than applying each deletion separately.
func deleteMatchesInBatch(filename string, deletions []models.Deletion) int {
	if isFileBlocked(filename) {
		return 0
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	filtered := []string{}
	count := 0

	for _, line := range lines {
		shouldDelete := false

		for _, deletion := range deletions {
			if deletion.Pattern.MatchString(line) {
				shouldDelete = true
				count++
				break
			}
		}

		if !shouldDelete {
			filtered = append(filtered, line)
		}
	}

	if count > 0 {
		os.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
	}

	return count
}
