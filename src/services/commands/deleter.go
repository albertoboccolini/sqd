package commands

import (
	"regexp"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Deleter struct {
	processor *files.Processor
	utils     *services.Utils
}

func NewDeleter(processor *files.Processor, utils *services.Utils) *Deleter {
	return &Deleter{
		processor: processor,
		utils:     utils,
	}
}

func (deleter *Deleter) Single(file string, pattern *regexp.Regexp, negate bool) (int, error) {
	return deleter.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		filtered := []string{}
		count := 0
		for _, line := range lines {
			matches := pattern.MatchString(line)
			if negate {
				matches = !matches
			}

			if !matches {
				filtered = append(filtered, line)
				continue
			}

			count++
		}
		return filtered, count
	})
}

func (deleter *Deleter) Batch(file string, deletions []models.Deletion) (int, error) {
	return deleter.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		filtered := []string{}
		count := 0
		for _, line := range lines {
			shouldDelete := false
			for _, deletion := range deletions {
				matches := deletion.Pattern.MatchString(line)
				if deletion.Negate {
					matches = !matches
				}

				if matches {
					shouldDelete = true
					break
				}
			}

			if shouldDelete {
				count++
				continue
			}

			filtered = append(filtered, line)
		}
		return filtered, count
	})
}
