package dry_mode

import (
	"fmt"
	"regexp"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
)

type ChangeProcessor struct {
	fileReader   *FileReader
	utils        *services.Utils
	printEnabled bool
}

func NewChangeProcessor(fileReader *FileReader, utils *services.Utils) *ChangeProcessor {
	return &ChangeProcessor{
		fileReader: fileReader,
		utils:      utils,
	}
}

func (changeProcessor *ChangeProcessor) printUpdateMessage(lineNumber int, file string, line string) {
	fmt.Printf("Updated line %d on %s content: %s\n", lineNumber+1, file, line)
}

func (changeProcessor *ChangeProcessor) printDeletionMessage(lineNumber int, file string, line string) {
	fmt.Printf("Deleted line %d on %s content: %s\n", lineNumber+1, file, line)
}

func (changeProcessor *ChangeProcessor) processSingleUpdate(lines []string, file string, pattern *regexp.Regexp, negate bool, replace string) int {
	updateCount := 0

	for lineIndex, originalLine := range lines {
		matchesPattern := pattern.MatchString(originalLine)
		shouldProcess := matchesPattern != negate

		if !shouldProcess {
			continue
		}

		modifiedLine := replace
		if !negate {
			modifiedLine = pattern.ReplaceAllLiteralString(originalLine, replace)
		}

		lineChanged := modifiedLine != originalLine
		if !lineChanged {
			continue
		}

		updateCount++

		if changeProcessor.printEnabled {
			changeProcessor.printUpdateMessage(lineIndex, file, modifiedLine)
		}
	}

	return updateCount
}

func (changeProcessor *ChangeProcessor) processBatchUpdates(lines []string, file string, replacements []models.Replacement) int {
	updateCount := 0

	for lineIndex, originalLine := range lines {
		currentLine := originalLine
		lineProcessed := false

		for _, replacement := range replacements {
			matchesPattern := replacement.Pattern.MatchString(currentLine)
			shouldProcess := matchesPattern != replacement.Negate

			if !shouldProcess {
				continue
			}

			currentLine = replacement.Replace
			if !replacement.Negate {
				currentLine = replacement.Pattern.ReplaceAllLiteralString(currentLine, replacement.Replace)
			}

			lineProcessed = true
			break
		}

		lineChanged := currentLine != originalLine
		if !lineProcessed || !lineChanged {
			continue
		}

		updateCount++

		if changeProcessor.printEnabled {
			changeProcessor.printUpdateMessage(lineIndex, file, currentLine)
		}
	}

	return updateCount
}

func (changeProcessor *ChangeProcessor) processUpdates(lines []string, file string, command models.Command) int {
	if command.IsBatch {
		return changeProcessor.processBatchUpdates(lines, file, command.Replacements)
	}
	return changeProcessor.processSingleUpdate(lines, file, command.Pattern, command.NegateContent, command.Replace)
}

func (changeProcessor *ChangeProcessor) processSingleDeletion(lines []string, file string, pattern *regexp.Regexp, negate bool) int {
	deletionCount := 0

	for lineIndex, line := range lines {
		matchesPattern := pattern.MatchString(line)
		shouldDelete := matchesPattern != negate

		if !shouldDelete {
			continue
		}

		deletionCount++

		if changeProcessor.printEnabled {
			changeProcessor.printDeletionMessage(lineIndex, file, line)
		}
	}

	return deletionCount
}

func (changeProcessor *ChangeProcessor) processBatchDeletions(lines []string, file string, deletions []models.Deletion) int {
	deletionCount := 0

	for lineIndex, line := range lines {
		lineDeleted := false

		for _, deletion := range deletions {
			matchesPattern := deletion.Pattern.MatchString(line)
			shouldDelete := matchesPattern != deletion.Negate

			if !shouldDelete {
				continue
			}

			deletionCount++
			lineDeleted = true

			if changeProcessor.printEnabled {
				changeProcessor.printDeletionMessage(lineIndex, file, line)
			}

			break
		}

		if !lineDeleted {
			continue
		}
	}

	return deletionCount
}

func (changeProcessor *ChangeProcessor) processDeletions(lines []string, file string, command models.Command) int {
	if command.IsBatch {
		return changeProcessor.processBatchDeletions(lines, file, command.Deletions)
	}

	return changeProcessor.processSingleDeletion(lines, file, command.Pattern, command.NegateContent)
}

func (changeProcessor *ChangeProcessor) WithPrinting() *ChangeProcessor {
	changeProcessor.printEnabled = true
	return changeProcessor
}

func (changeProcessor *ChangeProcessor) ProcessCommand(file string, command models.Command, stats *models.ExecutionStats) (int, error) {
	lines, err := changeProcessor.fileReader.ValidateAndReadFile(file)
	if err != nil {
		stats.Skipped++
		return 0, err
	}

	switch command.Action {
	case models.UPDATE:
		return changeProcessor.processUpdates(lines, file, command), nil
	case models.DELETE:
		return changeProcessor.processDeletions(lines, file, command), nil
	default:
		return 0, fmt.Errorf("unhandled command action: %v", command.Action)
	}
}
