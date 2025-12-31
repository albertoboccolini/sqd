package services

import (
	"fmt"
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type DryRunner struct {
	fileOperations FileOperations
	utils          *Utils
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

func (dryRunner *DryRunner) countUpdates(lines []string, command models.Command) int {
	count := 0
	for _, line := range lines {
		original := line

		if command.IsBatch {
			for _, replacement := range command.Replacements {
				if replacement.Pattern.MatchString(line) {
					line = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
					break
				}
			}
		} else if command.Pattern.MatchString(line) {
			line = command.Pattern.ReplaceAllLiteralString(line, command.Replace)
		}

		if line != original {
			count++
		}
	}

	return count
}

func (dryRunner *DryRunner) countDeletions(lines []string, command models.Command) int {
	count := 0
	for _, line := range lines {
		if command.IsBatch {
			for _, deletion := range command.Deletions {
				if deletion.Pattern.MatchString(line) {
					count++
					break
				}
			}
		} else if command.Pattern.MatchString(line) {
			count++
		}
	}

	return count
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
