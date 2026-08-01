package commands

import (
	"regexp"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Updater struct {
	processor *files.Processor
	utils     *services.Utils
}

func NewUpdater(processor *files.Processor, utils *services.Utils) *Updater {
	return &Updater{
		processor: processor,
		utils:     utils,
	}
}

func (updater *Updater) Single(file string, pattern *regexp.Regexp, negate bool, replace string) (int, error) {
	return updater.processor.ProcessFile(file, func(lines []string) ([]string, int) {
		count := 0
		for i, line := range lines {
			matches := pattern.MatchString(line)
			if negate {
				matches = !matches
			}

			if matches {
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
				matches := replacement.Pattern.MatchString(line)
				if replacement.Negate {
					matches = !matches
				}

				if matches {
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
