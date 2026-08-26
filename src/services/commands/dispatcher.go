package commands

import (
	"fmt"
	"time"

	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/dry_mode"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

type Dispatcher struct {
	searcher      *Searcher
	counter       *Counter
	updater       *Updater
	deleter       *Deleter
	transactioner *Transactioner
	dryModeRunner *dry_mode.Runner
	utils         *services.Utils
	parallelizer  *files.Parallelizer
}

func NewDispatcher(
	searcher *Searcher,
	counter *Counter,
	updater *Updater,
	deleter *Deleter,
	transactioner *Transactioner,
	dryModeRunner *dry_mode.Runner,
	utils *services.Utils,
	parallelizer *files.Parallelizer,
) *Dispatcher {
	return &Dispatcher{
		searcher:      searcher,
		counter:       counter,
		updater:       updater,
		deleter:       deleter,
		transactioner: transactioner,
		dryModeRunner: dryModeRunner,
		utils:         utils,
		parallelizer:  parallelizer,
	}
}

func (dispatcher *Dispatcher) printCount(total int, outputFormat models.OutputFormat) {
	switch outputFormat {
	case models.JSONOutput:
		fmt.Printf("{\"matches\":%d}\n", total)
	case models.CSVOutput:
		fmt.Println("matches")
		fmt.Printf("%d\n", total)
	default:
		fmt.Printf("%d matches\n", total)
	}
}

func (dispatcher *Dispatcher) Execute(command models.Command, files []string, useTransaction bool, dryRun bool, showDetailedOutputInDryMode bool, outputFormat models.OutputFormat) error {
	stats := models.ExecutionStats{StartTime: time.Now()}

	if (command.Action == models.UPDATE || command.Action == models.DELETE) &&
		command.WhereTarget == models.NAME {
		if command.Action == models.UPDATE {
			return displayable_errors.NewInvalidUpdateError("UPDATE operations cannot filter by file name. Use WHERE content = ... instead")
		}
		return displayable_errors.NewInvalidDeleteError("DELETE operations cannot filter by file name. Use WHERE content = ... instead")
	}

	if command.Pattern == nil &&
		command.WherePattern == nil &&
		((command.Action == models.SELECT ||
			command.Action == models.COUNT ||
			command.Action == models.UPDATE ||
			command.Action == models.DELETE) && !command.IsBatch) {
		return displayable_errors.NewInvalidWhereClauseError("Invalid query pattern")
	}

	if command.Action == models.UPDATE && !command.IsBatch && command.Replace == "" {
		return displayable_errors.NewInvalidUpdateError("Invalid replacement value")
	}

	if command.Action == models.COUNT {
		total, stats, err := dispatcher.counter.Count(files, command)
		dispatcher.printCount(total, outputFormat)
		if outputFormat == models.TextOutput {
			dispatcher.utils.PrintStats(stats)
		}

		return err
	}

	if command.Action == models.SELECT {
		stats, err := dispatcher.searcher.Select(files, command, outputFormat)
		if outputFormat == models.TextOutput {
			dispatcher.utils.PrintStats(stats)
		}

		return err
	}

	if command.Action == models.UPDATE {
		return dispatcher.runUpdate(command, files, dryRun, useTransaction, showDetailedOutputInDryMode, outputFormat, &stats)
	}

	if command.Action == models.DELETE {
		return dispatcher.runDelete(command, files, dryRun, useTransaction, showDetailedOutputInDryMode, outputFormat, &stats)
	}

	return fmt.Errorf("unhandled command action: %v", command.Action)
}

func (dispatcher *Dispatcher) runUpdate(command models.Command, files []string, dryRun, useTransaction, showDetailedOutputInDryMode bool, outputFormat models.OutputFormat, stats *models.ExecutionStats) error {
	if dryRun {
		err := dispatcher.dryModeRunner.Validate(command, files, stats, useTransaction, showDetailedOutputInDryMode)
		if err != nil && useTransaction {
			fmt.Println("Dry run: fail")
			return err
		}

		fmt.Println("Dry run: pass")
		return err
	}

	total, err := dispatcher.applyUpdate(command, files, useTransaction, stats)
	if err != nil {
		return err
	}

	dispatcher.utils.PrintUpdateMessage(total)
	if outputFormat == models.TextOutput {
		dispatcher.utils.PrintStats(*stats)
	}

	return nil
}

func (dispatcher *Dispatcher) applyUpdate(command models.Command, files []string, useTransaction bool, stats *models.ExecutionStats) (int, error) {
	updateFile := func(file string) (int, error) {
		if command.IsBatch {
			return dispatcher.updater.Batch(file, command.Replacements)
		}
		return dispatcher.updater.Single(file, command.Pattern, command.NegateContent, command.Replace)
	}

	if useTransaction {
		return dispatcher.transactioner.Update(files, updateFile, stats)
	}

	return dispatcher.collectMutationResults(files, updateFile, stats)
}

func (dispatcher *Dispatcher) runDelete(command models.Command, files []string, dryRun, useTransaction, showDetailedOutputInDryMode bool, outputFormat models.OutputFormat, stats *models.ExecutionStats) error {
	if dryRun {
		err := dispatcher.dryModeRunner.Validate(command, files, stats, useTransaction, showDetailedOutputInDryMode)
		if err != nil && useTransaction {
			fmt.Println("Dry run: fail")
			return err
		}
		fmt.Println("Dry run: pass")
		return err
	}

	total, err := dispatcher.applyDelete(command, files, useTransaction, stats)
	if err != nil {
		return err
	}

	dispatcher.utils.PrintDeleteMessage(total)
	if outputFormat == models.TextOutput {
		dispatcher.utils.PrintStats(*stats)
	}

	return nil
}

func (dispatcher *Dispatcher) applyDelete(command models.Command, files []string, useTransaction bool, stats *models.ExecutionStats) (int, error) {
	deleteFile := func(file string) (int, error) {
		if command.IsBatch {
			return dispatcher.deleter.Batch(file, command.Deletions)
		}

		return dispatcher.deleter.Single(file, command.Pattern, command.NegateContent)
	}

	if useTransaction {
		return dispatcher.transactioner.Delete(files, deleteFile, stats)
	}

	return dispatcher.collectMutationResults(files, deleteFile, stats)
}

func (dispatcher *Dispatcher) collectMutationResults(files []string, mutate func(string) (int, error), stats *models.ExecutionStats) (int, error) {
	errorCollection := models.NewErrorCollection()
	total := 0
	for _, file := range files {
		count, err := mutate(file)
		if err != nil {
			errorCollection.Add(err)
			stats.Skipped++
			continue
		}

		total += count
		stats.Processed++
	}

	if errorCollection.HasErrors() {
		return total, errorCollection
	}

	return total, nil
}
