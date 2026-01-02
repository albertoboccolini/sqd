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
	fileOperator   *FileOperator
}

func NewDryRunner(fileOperations FileOperations, utils *Utils) *DryRunner {
	return &DryRunner{fileOperations: fileOperations, utils: utils}
}

func (dryRunner *DryRunner) SetFileOperator(fileOperator *FileOperator) {
	dryRunner.fileOperator = fileOperator
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
	if command.IsBatch {
		return dryRunner.fileOperator.countUpdatesInLinesInBatch(lines, command.Replacements)
	}
	return dryRunner.fileOperator.countUpdatesInLines(lines, command.Pattern, command.Replace)
}

func (dryRunner *DryRunner) countDeletions(lines []string, command models.Command) int {
	if command.IsBatch {
		return dryRunner.fileOperator.countDeletionsInLinesInBatch(lines, command.Deletions)
	}
	return dryRunner.fileOperator.countDeletionsInLines(lines, command.Pattern)
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
