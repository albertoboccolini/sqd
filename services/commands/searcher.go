package commands

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/services"
	"github.com/overthinkinglabs/sqd/services/files"
)

type Searcher struct {
	parallelizer *files.Parallelizer
	sorter       *Sorter
	utils        *services.Utils
}

type searchResult struct {
	filePath    string
	lineNumber  int
	lineContent string
}

type fileResults struct {
	results  []searchResult
	hasMatch bool
}

func NewSearcher(parallelizer *files.Parallelizer, sorter *Sorter, utils *services.Utils) *Searcher {
	return &Searcher{
		parallelizer: parallelizer,
		sorter:       sorter,
		utils:        utils,
	}
}

func (searcher *Searcher) filterFilesByName(files []string, pattern *regexp.Regexp, negate bool) []string {
	filtered := make([]string, 0, len(files))
	for _, file := range files {
		fileName := filepath.Base(file)
		matches := pattern.MatchString(fileName)
		if negate {
			matches = !matches
		}

		if matches {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func matchesCondition(line string, pattern *regexp.Regexp, negate bool) bool {
	if pattern == nil {
		return true
	}

	matches := pattern.MatchString(line)
	if negate {
		matches = !matches
	}

	return matches
}

func matchesContent(line string, command models.Command) bool {
	if len(command.Substrings) == 1 {
		if !strings.Contains(line, command.Substrings[0]) {
			return false
		}

		if command.NegateContent {
			return false
		}
	} else {
		if !matchesCondition(line, command.Pattern, command.NegateContent) {
			return false
		}
	}

	if len(command.ExtraSubstrings) == 1 {
		if !strings.Contains(line, command.ExtraSubstrings[0]) {
			return false
		}

		if command.ExtraNegate {
			return false
		}
	} else {
		return matchesCondition(line, command.ExtraPattern, command.ExtraNegate)
	}

	return true
}

func countMatchingLines(file string, command models.Command) (int, error) {
	count := 0

	err := readFileLines(file, func(line string, lineNumber int) error {
		if matchesContent(line, command) {
			count++
		}

		return nil
	})

	return count, err
}

func (searcher *Searcher) Select(files []string, command models.Command) models.ExecutionStats {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	if command.SelectTarget == models.NAME && command.WhereTarget == models.NAME {
		results := make([]searchResult, 0, len(files))
		for _, file := range files {
			results = append(results, searchResult{filePath: file})
		}

		searcher.sorter.sortResults(results, command.OrderBy)
		for _, result := range results {
			fmt.Printf("%s\n", searcher.utils.HighlightName(result.filePath, command.WherePattern))
		}

		stats.Processed = len(files)
		return stats
	}

	if command.WhereTarget == models.NAME && (command.SelectTarget == models.ASTERISK || command.SelectTarget == models.CONTENT) {
		results := make([]searchResult, 0)
		for _, file := range files {
			err := readFileLines(file, func(line string, lineNumber int) error {
				results = append(results, searchResult{
					filePath:    file,
					lineNumber:  lineNumber,
					lineContent: line,
				})
				return nil
			})
			if err != nil {
				stats.Skipped++
				continue
			}

			stats.Processed++
		}

		searcher.sorter.sortResults(results, command.OrderBy)
		for _, result := range results {
			switch command.SelectTarget {
			case models.CONTENT:
				fmt.Printf("%s\n", result.lineContent)
			case models.ASTERISK:
				fmt.Printf("%s:%d: %s\n", searcher.utils.HighlightName(result.filePath, command.WherePattern), result.lineNumber, result.lineContent)
			}
		}

		return stats
	}

	allFileResults := make([]fileResults, len(files))

	searcher.parallelizer.ProcessFilesInParallelWithIndex(files, func(index int, file string) error {
		searchResults := fileResults{results: make([]searchResult, 0)}

		err := readFileLines(file, func(line string, lineNumber int) error {
			if matchesContent(line, command) {
				searchResults.hasMatch = true
				searchResults.results = append(searchResults.results, searchResult{
					filePath:    file,
					lineNumber:  lineNumber,
					lineContent: line,
				})
			}

			return nil
		})
		if err != nil {
			return err
		}

		allFileResults[index] = searchResults
		return nil
	}, &stats)

	results := make([]searchResult, 0)
	filesWithMatches := make([]string, 0)
	for i, searchResults := range allFileResults {
		if searchResults.hasMatch {
			results = append(results, searchResults.results...)
			if command.SelectTarget == models.NAME {
				filesWithMatches = append(filesWithMatches, files[i])
			}
		}
	}

	if command.SelectTarget == models.NAME {
		nameResults := make([]searchResult, 0, len(filesWithMatches))
		for _, file := range filesWithMatches {
			nameResults = append(nameResults, searchResult{filePath: file})
		}

		searcher.sorter.sortResults(nameResults, command.OrderBy)
		for _, result := range nameResults {
			fmt.Printf("%s\n", searcher.utils.HighlightName(result.filePath, command.Pattern))
		}
		return stats
	}

	searcher.sorter.sortResults(results, command.OrderBy)
	for _, result := range results {
		switch command.SelectTarget {
		case models.CONTENT:
			fmt.Printf("%s\n", searcher.utils.HighlightMatch(result.lineContent, command.Pattern))
		case models.ASTERISK:
			fmt.Printf("%s:%d: %s\n", result.filePath, result.lineNumber, searcher.utils.HighlightMatch(result.lineContent, command.Pattern))
		}
	}

	return stats
}
