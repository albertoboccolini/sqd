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
	lineReader   *LineReader
}

func NewCounter(parallelizer *files.Parallelizer, searcher *Searcher, lineReader *LineReader) *Counter {
	return &Counter{
		parallelizer: parallelizer,
		searcher:     searcher,
		lineReader:   lineReader,
	}
}

func (counter *Counter) Count(files []string, command models.Command) (int, models.ExecutionStats, error) {
	stats := models.ExecutionStats{StartTime: time.Now()}
	errorCollection := models.NewErrorCollection()

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = counter.searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	if len(command.SelectTargets) == 1 && command.SelectTargets[0] == models.NAME {
		if command.WhereTarget == models.NAME && command.WherePattern != nil {
			stats.Processed = len(files)
			return len(files), stats, nil
		}

		total, readErrors := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			found := false

			err := counter.lineReader.ReadFileLines(file, func(line string, lineNumber int) error {
				if counter.searcher.matchesContent(line, command) {
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

		counter.lineReader.MergeErrors(errorCollection, readErrors)
		return total, stats, counter.lineReader.ErrorCollectionOrNil(errorCollection)
	}

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		total := 0
		for _, file := range files {
			lineCount := 0

			err := counter.lineReader.ReadFileLines(file, func(line string, lineNumber int) error {
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

		return total, stats, counter.lineReader.ErrorCollectionOrNil(errorCollection)
	}

	total, readErrors := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
		return counter.searcher.countMatchingLines(file, command)
	}, &stats)

	counter.lineReader.MergeErrors(errorCollection, readErrors)
	return total, stats, counter.lineReader.ErrorCollectionOrNil(errorCollection)
}
