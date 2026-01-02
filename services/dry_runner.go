package services

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type DryRunner struct {
	fileOperations FileOperations
	utils          *Utils
	fileOperator   *FileOperator
}

func NewDryRunner(fileOperations FileOperations, utils *Utils) *DryRunner {
	return &DryRunner{fileOperations: fileOperations, utils: utils}
}

func (dryRunner *DryRunner) Validate(command models.Command, files []string, stats *models.ExecutionStats, useTransaction bool) bool {
	total := 0

	for _, file := range files {
		count, ok := dryRunner.validateAndCount(file, command, stats)
		if !ok {
			if useTransaction {
				return false
			}

			continue
		}

		total += count
		stats.Processed++
	}

	if command.Action == models.UPDATE {
		dryRunner.utils.PrintUpdateMessage(total)
	} else {
		fmt.Printf("Deleted: %d lines\n", total)
	}

	dryRunner.utils.PrintStats(*stats)
	return true
}

func (dryRunner *DryRunner) validateAndCount(file string, command models.Command, stats *models.ExecutionStats) (int, bool) {
	lines, ok := dryRunner.validateAndReadFile(file, stats)
	if !ok {
		return 0, false
	}

	if command.Action == models.UPDATE {
		return dryRunner.countUpdates(lines, command), true
	}

	return dryRunner.countDeletions(lines, command), true
}

func (dryRunner *DryRunner) countUpdatesInLines(lines []string, pattern *regexp.Regexp, replace string) int {
	count := 0
	for _, line := range lines {
		if pattern.MatchString(line) {
			newLine := pattern.ReplaceAllLiteralString(line, replace)
			if newLine != line {
				count++
			}
		}
	}

	return count
}

func (dryRunner *DryRunner) countUpdatesInLinesInBatch(lines []string, replacements []models.Replacement) int {
	count := 0
	for _, line := range lines {
		original := line
		for _, replacement := range replacements {
			if replacement.Pattern.MatchString(line) {
				line = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
				break
			}
		}

		if line != original {
			count++
		}
	}

	return count
}

func (dryRunner *DryRunner) countDeletionsInLines(lines []string, pattern *regexp.Regexp) int {
	count := 0
	for _, line := range lines {
		if pattern.MatchString(line) {
			count++
		}
	}

	return count
}

func (dryRunner *DryRunner) countDeletionsInLinesInBatch(lines []string, deletions []models.Deletion) int {
	count := 0
	for _, line := range lines {
		for _, deletion := range deletions {
			if deletion.Pattern.MatchString(line) {
				count++
				break
			}
		}
	}

	return count
}

func (dryRunner *DryRunner) countUpdates(lines []string, command models.Command) int {
	if command.IsBatch {
		return dryRunner.countUpdatesInLinesInBatch(lines, command.Replacements)
	}

	return dryRunner.countUpdatesInLines(lines, command.Pattern, command.Replace)
}

func (dryRunner *DryRunner) countDeletions(lines []string, command models.Command) int {
	if command.IsBatch {
		return dryRunner.countDeletionsInLinesInBatch(lines, command.Deletions)
	}

	return dryRunner.countDeletionsInLines(lines, command.Pattern)
}

func (dryRunner *DryRunner) validateAndReadFile(file string, stats *models.ExecutionStats) ([]string, bool) {
	if !dryRunner.utils.IsPathInsideCwd(file) {
		dryRunner.fail("invalid path: "+file, stats)
		return nil, false
	}

	if !dryRunner.utils.CanWriteFile(file) {
		dryRunner.fail("permission denied: "+file, stats)
		return nil, false
	}

	data, err := dryRunner.fileOperations.ReadFile(file)
	if err != nil {
		dryRunner.fail(err.Error(), stats)
		return nil, false
	}

	return strings.Split(string(data), "\n"), true
}

func (dryRunner *DryRunner) fail(msg string, stats *models.ExecutionStats) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	stats.Skipped++
}
