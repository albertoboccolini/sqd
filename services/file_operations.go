package services

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sqd/models"
)

func ExecuteCommand(cmd models.Command, files []string) {
	if cmd.Action == "COUNT" {
		total := 0
		for _, f := range files {
			total += countMatches(f, cmd.Pattern)
		}
		fmt.Printf("%d lines matched\n", total)
		return
	}

	if cmd.Action == "SELECT" {
		for _, f := range files {
			selectMatches(f, cmd.Pattern)
		}
		return
	}

	if cmd.Action == "UPDATE" {
		total := 0

		if cmd.IsBatch {
			for _, f := range files {
				total += updateFileBatch(f, cmd.Replacements)
			}
		}

		if !cmd.IsBatch {
			for _, f := range files {
				total += updateFile(f, cmd.Pattern, cmd.Replace)
			}
		}

		fmt.Printf("Updated: %d occurrences\n", total)
		return
	}

	if cmd.Action == "DELETE" {
		total := 0
		for _, f := range files {
			total += deleteMatches(f, cmd.Pattern)
		}
		fmt.Printf("Deleted: %d lines\n", total)
		return
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

func updateFileBatch(filename string, replacements []models.Replacement) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for i, line := range lines {
		for _, repl := range replacements {
			if repl.Pattern.MatchString(line) {
				lines[i] = repl.Pattern.ReplaceAllLiteralString(line, repl.Replace)
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
