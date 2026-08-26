package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

func displayName(file string, pattern string) string {
	if !strings.Contains(pattern, "/") {
		return file
	}

	lastSlash := strings.LastIndex(pattern, "/")
	baseDir := pattern[:lastSlash]
	if baseDir == "" {
		return file
	}

	relative, err := filepath.Rel(baseDir, file)
	if err != nil {
		return file
	}

	return relative
}

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

func (searcher *Searcher) limitResults(results []searchResult, limit int) []searchResult {
	if limit <= 0 || limit >= len(results) {
		return results
	}

	return results[:limit]
}

func (searcher *Searcher) selectTargets(command models.Command) []models.TokenType {
	if len(command.SelectTargets) > 0 {
		return command.SelectTargets
	}

	return []models.TokenType{command.SelectTarget}
}

func (searcher *Searcher) effectiveTargets(targets []models.TokenType) []models.TokenType {
	if len(targets) == 1 && targets[0] == models.ASTERISK {
		return []models.TokenType{models.NAME, models.LINE, models.CONTENT}
	}

	return targets
}

func (searcher *Searcher) includesTarget(targets []models.TokenType, target models.TokenType) bool {
	return slices.Contains(targets, target)
}

func (searcher *Searcher) printJSONResults(results []searchResult, command models.Command) {
	targets := searcher.effectiveTargets(searcher.selectTargets(command))
	records := make([]map[string]any, 0, len(results))

	for _, result := range results {
		record := make(map[string]any)

		for _, target := range targets {
			switch target {
			case models.NAME:
				record["name"] = displayName(result.filePath, command.File)
			case models.LINE:
				record["line"] = result.lineNumber
			case models.CONTENT:
				record["content"] = result.lineContent
			}
		}

		records = append(records, record)
	}

	data, _ := json.Marshal(records)
	fmt.Printf("%s\n", data)
}

func (searcher *Searcher) printCSVResults(results []searchResult, command models.Command) {
	targets := searcher.effectiveTargets(searcher.selectTargets(command))
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	header := make([]string, 0, len(targets))
	for _, target := range targets {
		switch target {
		case models.NAME:
			header = append(header, "name")
		case models.LINE:
			header = append(header, "line")
		case models.CONTENT:
			header = append(header, "content")
		}
	}

	_ = writer.Write(header)

	for _, result := range results {
		row := make([]string, 0, len(header))

		for _, target := range targets {
			switch target {
			case models.NAME:
				row = append(row, displayName(result.filePath, command.File))
			case models.LINE:
				row = append(row, fmt.Sprintf("%d", result.lineNumber))
			case models.CONTENT:
				row = append(row, result.lineContent)
			}
		}

		_ = writer.Write(row)
	}
}

func (searcher *Searcher) printStructuredResults(results []searchResult, command models.Command, outputFormat models.OutputFormat) {
	if outputFormat == models.JSONOutput {
		searcher.printJSONResults(results, command)
		return
	}

	if outputFormat == models.CSVOutput {
		searcher.printCSVResults(results, command)
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
	if !matchesCondition(line, command.Pattern, command.NegateContent) {
		return false
	}

	return matchesCondition(line, command.ExtraPattern, command.ExtraNegate)
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

func (searcher *Searcher) Select(files []string, command models.Command, outputFormat models.OutputFormat) (models.ExecutionStats, error) {
	stats := models.ExecutionStats{StartTime: time.Now()}
	errorCollection := models.NewErrorCollection()

	if command.WhereTarget == models.NAME && command.WherePattern != nil {
		files = searcher.filterFilesByName(files, command.WherePattern, command.NegateFileName)
	}

	targets := searcher.selectTargets(command)
	includeName := searcher.includesTarget(targets, models.NAME)
	includeLine := searcher.includesTarget(targets, models.LINE)
	includeContent := searcher.includesTarget(targets, models.CONTENT)
	includeAll := includeName && includeContent && includeLine

	if includeName && !includeContent && !includeLine && command.WhereTarget == models.NAME {
		results := make([]searchResult, 0, len(files))
		for _, file := range files {
			results = append(results, searchResult{filePath: file})
		}

		searcher.sorter.sortResults(results, command.OrderBy)
		results = searcher.limitResults(results, command.Limit)

		if outputFormat != models.TextOutput {
			searcher.printStructuredResults(results, command, outputFormat)
			stats.Processed = len(files)
			return stats, nil
		}

		for _, result := range results {
			fmt.Printf("%s\n", searcher.utils.HighlightName(displayName(result.filePath, command.File), command.WherePattern))
		}

		stats.Processed = len(files)
		return stats, nil
	}

	if command.WhereTarget == models.NAME && (includeAll || includeContent || includeLine) {
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
				errorCollection.Add(err)
				stats.Skipped++
				continue
			}

			stats.Processed++
		}

		searcher.sorter.sortResults(results, command.OrderBy)
		results = searcher.limitResults(results, command.Limit)

		if outputFormat != models.TextOutput {
			searcher.printStructuredResults(results, command, outputFormat)
			return stats, errorCollectionOrNil(errorCollection)
		}

		searcher.printTextResults(results, command, searcher.selectTargets(command))
		return stats, errorCollectionOrNil(errorCollection)
	}

	allFileResults := make([]fileResults, len(files))

	readErrors := searcher.parallelizer.ProcessFilesInParallelWithIndex(files, func(index int, file string) error {
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

	mergeErrors(errorCollection, readErrors)

	results := make([]searchResult, 0)
	filesWithMatches := make([]string, 0)
	for i, searchResults := range allFileResults {
		if searchResults.hasMatch {
			results = append(results, searchResults.results...)
			if includeName && !includeContent && !includeLine {
				filesWithMatches = append(filesWithMatches, files[i])
			}
		}
	}

	if includeName && !includeContent && !includeLine {
		nameResults := make([]searchResult, 0, len(filesWithMatches))
		for _, file := range filesWithMatches {
			nameResults = append(nameResults, searchResult{filePath: file})
		}

		searcher.sorter.sortResults(nameResults, command.OrderBy)
		nameResults = searcher.limitResults(nameResults, command.Limit)

		if outputFormat != models.TextOutput {
			searcher.printStructuredResults(nameResults, command, outputFormat)
			return stats, errorCollectionOrNil(errorCollection)
		}

		for _, result := range nameResults {
			fmt.Printf("%s\n", searcher.utils.HighlightName(displayName(result.filePath, command.File), command.Pattern))
		}

		return stats, errorCollectionOrNil(errorCollection)
	}

	searcher.sorter.sortResults(results, command.OrderBy)
	results = searcher.limitResults(results, command.Limit)

	if outputFormat != models.TextOutput {
		searcher.printStructuredResults(results, command, outputFormat)
		return stats, errorCollectionOrNil(errorCollection)
	}

	searcher.printTextResults(results, command, searcher.selectTargets(command))
	return stats, errorCollectionOrNil(errorCollection)
}

func (searcher *Searcher) printTextResults(results []searchResult, command models.Command, targets []models.TokenType) {
	if len(targets) == 1 && targets[0] == models.ASTERISK {
		for _, result := range results {
			content := result.lineContent
			if command.Pattern != nil {
				content = searcher.utils.HighlightMatch(content, command.Pattern)
			}

			fmt.Printf("%s:%d: %s\n", displayName(result.filePath, command.File), result.lineNumber, content)
		}

		return
	}

	effectiveTargets := searcher.effectiveTargets(targets)

	for _, result := range results {
		var parts []string

		for _, target := range effectiveTargets {
			switch target {
			case models.NAME:
				parts = append(parts, searcher.utils.HighlightName(displayName(result.filePath, command.File), command.WherePattern))
			case models.LINE:
				parts = append(parts, fmt.Sprintf("%d", result.lineNumber))
			case models.CONTENT:
				content := result.lineContent
				if command.Pattern != nil {
					content = searcher.utils.HighlightMatch(content, command.Pattern)
				}

				parts = append(parts, content)
			}
		}

		fmt.Println(strings.Join(parts, "\t"))
	}
}
