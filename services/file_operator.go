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

func (osFileOperations *OSFileOperations) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (osFileOperations *OSFileOperations) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (osFileOperations *OSFileOperations) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

type FileOperator struct {
	fileOperations FileOperations
	utils          *Utils
	dryRunPreview  *DryRunner
}

func NewFileOperator(utils *Utils) *FileOperator {
	fileOperations := &OSFileOperations{}
	return &FileOperator{
		fileOperations: fileOperations,
		utils:          utils,
		dryRunPreview:  NewDryRunner(fileOperations, utils),
	}
}

func (fileOperator *FileOperator) executeDryRun(command models.Command, files []string, useTransaction bool, stats models.ExecutionStats) {
	isValid := fileOperator.dryRunPreview.Validate(command, files, &stats, useTransaction)
	status := "fail"
	if isValid {
		status = "pass"
	}
	fmt.Printf("Dry run: %s\n", status)
}

func (fileOperator *FileOperator) ExecuteCommand(command models.Command, files []string, useTransaction bool, dryRun bool) {
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
			count, err := fileOperator.countMatches(file, command.Pattern)
			if err != nil {
				fileOperator.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("%d lines matched\n", total)
		fileOperator.utils.PrintStats(stats)
		return
	}

	if command.Action == models.SELECT {
		for _, file := range files {
			err := fileOperator.selectMatches(file, command.Pattern)
			if err != nil {
				fileOperator.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			stats.Processed++
		}

		fileOperator.utils.PrintStats(stats)
		return
	}

	if command.Action == models.UPDATE {
		if dryRun {
			fileOperator.executeDryRun(command, files, useTransaction, stats)
			return
		}

		if useTransaction {
			fileOperator.executeUpdateTransaction(command, files, &stats)
			return
		}

		total := 0
		if command.IsBatch {
			for _, file := range files {
				count, err := fileOperator.updateFileInBatch(file, command.Replacements)
				if err != nil {
					fileOperator.utils.PrintProcessingErrorMessage(file, err)
					stats.Skipped++
					continue
				}
				total += count
				stats.Processed++
			}

			fileOperator.utils.PrintUpdateMessage(total)
			fileOperator.utils.PrintStats(stats)
			return
		}

		for _, file := range files {
			count, err := fileOperator.updateFile(file, command.Pattern, command.Replace)
			if err != nil {
				fileOperator.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fileOperator.utils.PrintUpdateMessage(total)
		fileOperator.utils.PrintStats(stats)
		return
	}

	if command.Action == models.DELETE {
		if dryRun {
			fileOperator.executeDryRun(command, files, useTransaction, stats)
			return
		}

		if useTransaction {
			fileOperator.executeDeleteTransaction(command, files, &stats)
			return
		}

		total := 0

		if command.IsBatch {
			for _, file := range files {
				count, err := fileOperator.deleteMatchesInBatch(file, command.Deletions)
				if err != nil {
					fileOperator.utils.PrintProcessingErrorMessage(file, err)
					stats.Skipped++
					continue
				}
				total += count
				stats.Processed++
			}

			fmt.Printf("Deleted: %d lines\n", total)
			fileOperator.utils.PrintStats(stats)
			return
		}

		for _, file := range files {
			count, err := fileOperator.deleteMatches(file, command.Pattern)
			if err != nil {
				fileOperator.utils.PrintProcessingErrorMessage(file, err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		fmt.Printf("Deleted: %d lines\n", total)
		fileOperator.utils.PrintStats(stats)
	}
}

func (fileOperator *FileOperator) countMatches(filename string, pattern *regexp.Regexp) (int, error) {
	data, err := fileOperator.fileOperations.ReadFile(filename)
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

func (fileOperator *FileOperator) selectMatches(filename string, pattern *regexp.Regexp) error {
	data, err := fileOperator.fileOperations.ReadFile(filename)
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

func (fileOperator *FileOperator) updateFile(filename string, pattern *regexp.Regexp, replace string) (int, error) {
	if !fileOperator.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !fileOperator.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := fileOperator.fileOperations.ReadFile(filename)
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
		err = fileOperator.fileOperations.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (fileOperator *FileOperator) updateFileInBatch(filename string, replacements []models.Replacement) (int, error) {
	if !fileOperator.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !fileOperator.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := fileOperator.fileOperations.ReadFile(filename)
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
		err = fileOperator.fileOperations.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (fileOperator *FileOperator) deleteMatches(filename string, pattern *regexp.Regexp) (int, error) {
	if !fileOperator.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !fileOperator.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := fileOperator.fileOperations.ReadFile(filename)
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
		err = fileOperator.fileOperations.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (fileOperator *FileOperator) deleteMatchesInBatch(filename string, deletions []models.Deletion) (int, error) {
	if !fileOperator.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}

	if !fileOperator.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	data, err := fileOperator.fileOperations.ReadFile(filename)
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
		err = fileOperator.fileOperations.WriteFile(filename, []byte(strings.Join(filtered, "\n")), 0644)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (fileOperator *FileOperator) checkFilesBeforeTransaction(files []string) {
	for _, file := range files {
		if !fileOperator.utils.IsPathInsideCwd(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: invalid path %s\n", file)
			os.Exit(1)
		}
		if !fileOperator.utils.CanWriteFile(file) {
			fmt.Fprintf(os.Stderr, "Transaction failed: cannot write %s\n", file)
			os.Exit(1)
		}
	}
}

func (fileOperator *FileOperator) executeUpdateTransaction(command models.Command, files []string, stats *models.ExecutionStats) {
	fileOperator.checkFilesBeforeTransaction(files)

	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := fileOperator.fileOperations.Rename(file, backupPath); err != nil {
			fileOperator.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		var count int
		var err error

		if command.IsBatch {
			count, err = fileOperator.updateFileInBatch(backupPath, command.Replacements)
		} else {
			count, err = fileOperator.updateFile(backupPath, command.Pattern, command.Replace)
		}

		if err != nil {
			fileOperator.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		if err := fileOperator.fileOperations.Rename(backupPath, file); err != nil {
			fileOperator.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		total += count
		stats.Processed++
	}

	fileOperator.utils.PrintUpdateMessage(total)
	fileOperator.utils.PrintStats(*stats)
}

func (fileOperator *FileOperator) executeDeleteTransaction(command models.Command, files []string, stats *models.ExecutionStats) {
	fileOperator.checkFilesBeforeTransaction(files)
	backups := make([]fileBackup, 0, len(files))
	total := 0

	for _, file := range files {
		backupPath := file + ".sqd_backup"
		if err := fileOperator.fileOperations.Rename(file, backupPath); err != nil {
			fileOperator.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}
		backups = append(backups, fileBackup{original: file, backup: backupPath})

		var count int
		var err error

		if command.IsBatch {
			count, err = fileOperator.deleteMatchesInBatch(backupPath, command.Deletions)
		} else {
			count, err = fileOperator.deleteMatches(backupPath, command.Pattern)
		}

		if err != nil {
			fileOperator.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		if err := fileOperator.fileOperations.Rename(backupPath, file); err != nil {
			fileOperator.rollbackFiles(backups)
			fmt.Fprintf(os.Stderr, "Transaction failed: %v\n", err)
			return
		}

		total += count
		stats.Processed++
	}

	fmt.Printf("Deleted: %d lines\n", total)
	fileOperator.utils.PrintStats(*stats)
}

func (fileOperator *FileOperator) rollbackFiles(backups []fileBackup) {
	for _, backup := range backups {
		if err := fileOperator.fileOperations.Rename(backup.backup, backup.original); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed for %s -> %s: %v\n", backup.backup, backup.original, err)
		}
	}
}
