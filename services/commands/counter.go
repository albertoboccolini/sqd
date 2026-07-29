package commands

import (
	"io"
	"time"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/services/files"
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

func (counter *Counter) Count(files []string, command models.Command) (int, models.ExecutionStats) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = counter.searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	switch command.SelectTarget {
	case models.NAME:
		if command.WhereTarget == models.NAME && command.WherePattern != nil {
			stats.Processed = len(files)
			return len(files), stats
		}

		total := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
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

		return total, stats
	case models.CONTENT, models.ASTERISK:
		if command.WhereTarget == models.NAME && command.WherePattern != nil {
			total := 0
			for _, file := range files {
				lineCount := 0

				err := readFileLines(file, func(line string, lineNumber int) error {
					lineCount++
					return nil
				})
				if err != nil {
					stats.Skipped++
					continue
				}

				total += lineCount
				stats.Processed++
			}
			return total, stats
		}

		total := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, command)
		}, &stats)

		return total, stats
	default:
		total := counter.parallelizer.ProcessFilesInParallel(files, func(file string) (int, error) {
			return countMatchingLines(file, command)
		}, &stats)

		return total, stats
	}
}
