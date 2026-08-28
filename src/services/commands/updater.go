package commands

import (
	"regexp"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Updater struct {
	processor  *files.Processor
	utils      *services.Utils
	lineReader *LineReader
}

func NewUpdater(processor *files.Processor, utils *services.Utils, lineReader *LineReader) *Updater {
	return &Updater{
		processor:  processor,
		utils:      utils,
		lineReader: lineReader,
	}
}

func (updater *Updater) Single(file string, pattern *regexp.Regexp, negate bool, replace string) (int, error) {
	return updater.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		count := 0
		for i, line := range lines {
			if updater.lineReader.MatchesCondition(line, pattern, negate) {
				lines[i] = replace
				if !negate {
					lines[i] = pattern.ReplaceAllLiteralString(line, replace)
				}

				count++
			}
		}
		return lines, count
	})
}

func (updater *Updater) Batch(file string, replacements []models.Replacement) (int, error) {
	return updater.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		count := 0
		for i, line := range lines {
			for _, replacement := range replacements {
				if updater.lineReader.MatchesCondition(line, replacement.Pattern, replacement.Negate) {
					lines[i] = replacement.Replace
					if !replacement.Negate {
						lines[i] = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
					}

					count++
					break
				}
			}
		}
		return lines, count
	})
}
