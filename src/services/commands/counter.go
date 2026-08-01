package commands

import (
	"io"
	"time"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Counter struct {
	parallelizer *files.Parallelizer
	searcher     *Searcher
}

func NewCounter(parallelizer *files.Parallelizer, searcher *Searcher) *Counter {
	return &Counter{
		parallelizer: parallelizer,
		searcher:     searcher,
	}
}

func (counter *Counter) Count(files []string, command models.Command) (int, models.ExecutionStats, error) {
	stats := models.ExecutionStats{StartTime: time.Now()}
	errorCollection := models.NewErrorCollection()

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = counter.searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	if command.SelectTarget == models.NAME {
		if command.WhereTarget == models.NAME && command.WherePattern != nil {
			stats.Processed = len(files)
			return len(files), stats, nil
		}

		total, readErrors := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			found := false

			err := readFileLines(file, func(line string, lineNumber int) error {
				if matchesContent(line, command) {
					found = true
					return io.EOF
				}

				return nil
			})
			if err != nil {
				return 0, err
			}

			if found {
				return 1, nil
			}

			return 0, nil
		}, &stats)

		mergeErrors(errorCollection, readErrors)
		return total, stats, errorCollectionOrNil(errorCollection)
	}

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		total := 0
		for _, file := range files {
			lineCount := 0

			err := readFileLines(file, func(line string, lineNumber int) error {
				lineCount++
				return nil
			})
			if err != nil {
				errorCollection.Add(err)
				stats.Skipped++
				continue
			}

			total += lineCount
			stats.Processed++
		}

		return total, stats, errorCollectionOrNil(errorCollection)
	}

	total, readErrors := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
		return countMatchingLines(file, command)
	}, &stats)

	mergeErrors(errorCollection, readErrors)
	return total, stats, errorCollectionOrNil(errorCollection)
}
