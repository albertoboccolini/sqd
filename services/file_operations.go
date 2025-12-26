package services

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/albertoboccolini/sqd/models"
)

type fileBackup struct {
	original string
	backup   string
}

func ExecuteCommand(command models.Command, files []string, useTransaction bool) {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if command.Pattern == nil && ((command.Action == models.SELECT ||
		command.Action == models.COUNT ||
		command.Action == models.UPDATE ||
		command.Action == models.DELETE) && !command.IsBatch) {
		fmt.Fprintf(os.Stderr, "Error: Invalid query pattern\n")
		return
	}

	if command.Action == models.UPDATE && !command.IsBatch && command.Replace == "" {
		fmt.Fprintf(os.Stderr, "Error: Invalid replacement value\n")
		return
	}

	if command.Action == models.COUNT {
		total := 0
		for _, file := range files {
			count, err := countMatches(file, command.Pattern)
			if err != nil {
				PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("%d lines matched\n", total)
		PrintStats(stats)
		return
	}

	if command.Action == models.SELECT {
		for _, file := range files {
			err := selectMatches(file, command.Pattern)
			if err != nil {
				PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			stats.Processed++
		}

		PrintStats(stats)
		return
	}

	if command.Action == models.UPDATE {
		if useTransaction {
			executeUpdateTransaction(command, files, &stats)
			return
		}

		total := 0
		if command.IsBatch {
			for _, file := range files {
				count, err := updateFileInBatch(file, command.Replacements)
				if err != nil {
					PrintProcessingErrorMessage(file, err)
					stats.Skipped++
					continue
				}
				total += count
				stats.Processed++
			}

			PrintUpdateMessage(total)
			PrintStats(stats)
			return
		}

		for _, file := range files {
			count, err := updateFile(file, command.Pattern, command.Replace)
			if err != nil {
				PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		PrintUpdateMessage(total)
		PrintStats(stats)
		return
	}

	if command.Action == models.DELETE {
		if useTransaction {
			executeDeleteTransaction(command, files, &stats)
			return
		}

		total := 0

		if command.IsBatch {
			for _, file := range files {
				count, err := deleteMatchesInBatch(file, command.Deletions)
				if err != nil {
					PrintProcessingErrorMessage(file, err)
					stats.Skipped++
					continue
				}
				total += count
				stats.Processed++
			}

			fmt.Printf("Deleted: %d lines\n", total)
			PrintStats(stats)
			return
		}

		for _, file := range files {
			count, err := deleteMatches(file, command.Pattern)
			if err != nil {
				PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("Deleted: %d lines\n", total)
		PrintStats(stats)
	}
}

func countMatches(filename string, pattern *regexp.Regexp) (int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for _, line := range lines {
		if pattern.MatchString(line) {
			count++
		}
	}

	return count, nil
}

func selectMatches(filename string, pattern *regexp.Regexp) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if pattern.MatchString(line) {
			fmt.Printf("%s:%d: %s\n", filename, i+1, line)
		}
	}

	return nil
}

func updateFile(filename string, pattern *regexp.Regexp, replace string) (int, error) {
	if !IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !canWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for i, line := range lines {
		if pattern.MatchString(line) {
			lines[i] = pattern.ReplaceAllLiteralString(line, replace)
			count++
		}
	}

	if count > 0 {
		err = os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

// updateFileInBatch applies multiple replacements to the file in a single pass.
// This is more efficient than applying each replacement separately.
func updateFileInBatch(filename string, replacements []models.Replacement) (int, error) {
	if !IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !canWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	count := 0

	for i, line := range lines {
		for _, replacement := range replacements {
			if replacement.Pattern.MatchString(line) {
				lines[i] = replacement.Pattern.ReplaceAllLiteralString(line, replacement.Replace)
				count++
				break
			}
		}
	}

	if count > 0 {
		err = os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func deleteMatches(filename string, pattern *regexp.Regexp) (int, error) {
	if !IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !canWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	filtered := []string{}
	count := 0

	for _, line := range lines {
		if !pattern.MatchString(line) {
			filtered = append(filtered, line)
			continue
		}
		count++
	}

	if count > 0 {
		err = os.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

// deleteMatchesInBatch applies multiple deletions to the file in a single pass.
// This is more efficient than applying each deletion separately.
func deleteMatchesInBatch(filename string, deletions []models.Deletion) (int, error) {
	if !IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !canWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	filtered := []string{}
	count := 0

	for _, line := range lines {
		shouldDelete := false

		for _, deletion := range deletions {
			if deletion.Pattern.MatchString(line) {
				shouldDelete = true
				count++
				break
			}
		}

		if !shouldDelete {
			filtered = append(filtered, line)
		}
	}

	if count > 0 {
		err = os.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func checkFilesBeforeTransaction(files []string) {
	for _, file := range files {
		if !IsPathInsideCwd(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: invalid path %s\n", file)
			os.Exit(1)
		}
		if !canWriteFile(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: cannot write %s\n", file)
			os.Exit(1)
		}
	}
}

func executeUpdateTransaction(command models.Command, files []string, stats *models.ExecutionStats) {
	checkFilesBeforeTransaction(files)

	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := os.Rename(file, backupPath); err != nil {
			rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		var count int
		var err error

		if command.IsBatch {
			count, err = updateFileInBatch(backupPath, command.Replacements)
		} else {
			count, err = updateFile(backupPath, command.Pattern, command.Replace)
		}

		if err != nil {
			rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		if err := os.Rename(backupPath, file); err != nil {
			rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		total += count
		stats.Processed++
	}

	PrintUpdateMessage(total)
	PrintStats(*stats)
}

func executeDeleteTransaction(command models.Command, files []string, stats *models.ExecutionStats) {
	checkFilesBeforeTransaction(files)

	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := os.Rename(file, backupPath); err != nil {
			rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		var count int
		var err error

		if command.IsBatch {
			count, err = deleteMatchesInBatch(backupPath, command.Deletions)
		} else {
			count, err = deleteMatches(backupPath, command.Pattern)
		}

		if err != nil {
			rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		if err := os.Rename(backupPath, file); err != nil {
			rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		total += count
		stats.Processed++
	}

	fmt.Printf("Deleted: %d lines\n", total)
	PrintStats(*stats)
}

func rollbackFiles(backups []fileBackup) {
	for _, backup := range backups {
		if err := os.Rename(backup.backup, backup.original); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed for %s -> %s: %v\n", backup.backup, backup.original, err)
		}
	}
}
