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

type FileOperations interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	Rename(oldpath, newpath string) error
}

type OSFileOperations struct{}

func (osfo *OSFileOperations) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (osfo *OSFileOperations) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (osfo *OSFileOperations) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

type FileOperator struct {
	fileOperations FileOperations
	utils          *Utils
}

func NewFileOperator(utils *Utils) *FileOperator {
	return &FileOperator{
		fileOperations: &OSFileOperations{},
		utils:          utils,
	}
}

func (executor *FileOperator) ExecuteCommand(command models.Command, files []string, useTransaction bool) {
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
			count, err := executor.countMatches(file, command.Pattern)
			if err != nil {
				executor.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("%d lines matched\n", total)
		executor.utils.PrintStats(stats)
		return
	}

	if command.Action == models.SELECT {
		for _, file := range files {
			err := executor.selectMatches(file, command.Pattern)
			if err != nil {
				executor.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			stats.Processed++
		}

		executor.utils.PrintStats(stats)
		return
	}

	if command.Action == models.UPDATE {
		if useTransaction {
			executor.executeUpdateTransaction(command, files, &stats)
			return
		}

		total := 0
		if command.IsBatch {
			for _, file := range files {
				count, err := executor.updateFileInBatch(file, command.Replacements)
				if err != nil {
					executor.utils.PrintProcessingErrorMessage(file, err)
					stats.Skipped++
					continue
				}
				total += count
				stats.Processed++
			}

			executor.utils.PrintUpdateMessage(total)
			executor.utils.PrintStats(stats)
			return
		}

		for _, file := range files {
			count, err := executor.updateFile(file, command.Pattern, command.Replace)
			if err != nil {
				executor.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		executor.utils.PrintUpdateMessage(total)
		executor.utils.PrintStats(stats)
		return
	}

	if command.Action == models.DELETE {
		if useTransaction {
			executor.executeDeleteTransaction(command, files, &stats)
			return
		}

		total := 0

		if command.IsBatch {
			for _, file := range files {
				count, err := executor.deleteMatchesInBatch(file, command.Deletions)
				if err != nil {
					executor.utils.PrintProcessingErrorMessage(file, err)
					stats.Skipped++
					continue
				}
				total += count
				stats.Processed++
			}

			fmt.Printf("Deleted: %d lines\n", total)
			executor.utils.PrintStats(stats)
			return
		}

		for _, file := range files {
			count, err := executor.deleteMatches(file, command.Pattern)
			if err != nil {
				executor.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("Deleted: %d lines\n", total)
		executor.utils.PrintStats(stats)
	}
}

func (executor *FileOperator) countMatches(filename string, pattern *regexp.Regexp) (int, error) {
	data, err := executor.fileOperations.ReadFile(filename)
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

func (executor *FileOperator) selectMatches(filename string, pattern *regexp.Regexp) error {
	data, err := executor.fileOperations.ReadFile(filename)
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

func (executor *FileOperator) updateFile(filename string, pattern *regexp.Regexp, replace string) (int, error) {
	if !executor.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !executor.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := executor.fileOperations.ReadFile(filename)
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
		err = executor.fileOperations.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (executor *FileOperator) updateFileInBatch(filename string, replacements []models.Replacement) (int, error) {
	if !executor.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !executor.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := executor.fileOperations.ReadFile(filename)
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
		err = executor.fileOperations.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (executor *FileOperator) deleteMatches(filename string, pattern *regexp.Regexp) (int, error) {
	if !executor.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !executor.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := executor.fileOperations.ReadFile(filename)
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
		err = executor.fileOperations.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (executor *FileOperator) deleteMatchesInBatch(filename string, deletions []models.Deletion) (int, error) {
	if !executor.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !executor.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := executor.fileOperations.ReadFile(filename)
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
		err = executor.fileOperations.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (executor *FileOperator) checkFilesBeforeTransaction(files []string) {
	for _, file := range files {
		if !executor.utils.IsPathInsideCwd(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: invalid path %s\n", file)
			os.Exit(1)
		}
		if !executor.utils.CanWriteFile(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: cannot write %s\n", file)
			os.Exit(1)
		}
	}
}

func (executor *FileOperator) executeUpdateTransaction(command models.Command, files []string, stats *models.ExecutionStats) {
	executor.checkFilesBeforeTransaction(files)

	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := executor.fileOperations.Rename(file, backupPath); err != nil {
			executor.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		var count int
		var err error

		if command.IsBatch {
			count, err = executor.updateFileInBatch(backupPath, command.Replacements)
		} else {
			count, err = executor.updateFile(backupPath, command.Pattern, command.Replace)
		}

		if err != nil {
			executor.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		if err := executor.fileOperations.Rename(backupPath, file); err != nil {
			executor.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		total += count
		stats.Processed++
	}

	executor.utils.PrintUpdateMessage(total)
	executor.utils.PrintStats(*stats)
}

func (executor *FileOperator) executeDeleteTransaction(command models.Command, files []string, stats *models.ExecutionStats) {
	executor.checkFilesBeforeTransaction(files)
	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := executor.fileOperations.Rename(file, backupPath); err != nil {
			executor.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		var count int
		var err error

		if command.IsBatch {
			count, err = executor.deleteMatchesInBatch(backupPath, command.Deletions)
		} else {
			count, err = executor.deleteMatches(backupPath, command.Pattern)
		}

		if err != nil {
			executor.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		if err := executor.fileOperations.Rename(backupPath, file); err != nil {
			executor.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		total += count
		stats.Processed++
	}

	fmt.Printf("Deleted: %d lines\n", total)
	executor.utils.PrintStats(*stats)
}

func (executor *FileOperator) rollbackFiles(backups []fileBackup) {
	for _, backup := range backups {
		if err := executor.fileOperations.Rename(backup.backup, backup.original); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed for %s -> %s: %v\n", backup.backup, backup.original, err)
		}
	}
}
