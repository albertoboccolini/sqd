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

func (dispatcher *Dispatcher) Execute(command models.Command, files []string, useTransaction bool, dryRun bool, showDetailedOutputInDryMode bool) error {
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
		fmt.Printf("%d matches\n", total)
		dispatcher.utils.PrintStats(stats)
		return err
	}

	if command.Action == models.SELECT {
		stats, err := dispatcher.searcher.Select(files, command)
		dispatcher.utils.PrintStats(stats)
		return err
	}

	if command.Action == models.UPDATE {
		if dryRun {
			err := dispatcher.dryModeRunner.Validate(command, files, &stats, useTransaction, showDetailedOutputInDryMode)
			if err != nil && useTransaction {
				fmt.Println("Dry run: fail")
				return err
			}

			fmt.Println("Dry run: pass")
			return err
		}

		if useTransaction {
			var updateFunc func(string) (int, error)
			if command.IsBatch {
				updateFunc = func(file string) (int, error) {
					return dispatcher.updater.Batch(file, command.Replacements)
				}
			} else {
				updateFunc = func(file string) (int, error) {
					return dispatcher.updater.Single(file, command.Pattern, command.NegateContent, command.Replace)
				}
			}

			total, err := dispatcher.transactioner.Update(files, updateFunc, &stats)
			if err != nil {
				return err
			}
			dispatcher.utils.PrintUpdateMessage(total)
			dispatcher.utils.PrintStats(stats)
			return nil
		}

		errorCollection := models.NewErrorCollection()
		total := 0
		for _, file := range files {
			var count int
			var err error

			if command.IsBatch {
				count, err = dispatcher.updater.Batch(file, command.Replacements)
			} else {
				count, err = dispatcher.updater.Single(file, command.Pattern, command.NegateContent, command.Replace)
			}

			if err != nil {
				errorCollection.Add(err)
				stats.Skipped++
				continue
			}
			total += count
			stats.Processed++
		}

		if errorCollection.HasErrors() {
			dispatcher.utils.PrintUpdateMessage(total)
			dispatcher.utils.PrintStats(stats)
			return errorCollection
		}

		dispatcher.utils.PrintUpdateMessage(total)
		dispatcher.utils.PrintStats(stats)
		return nil
	}

	if command.Action == models.DELETE {
		if dryRun {
			err := dispatcher.dryModeRunner.Validate(command, files, &stats, useTransaction, showDetailedOutputInDryMode)
			if err != nil && useTransaction {
				fmt.Println("Dry run: fail")
				return err
			}
			fmt.Println("Dry run: pass")
			return err
		}

		if useTransaction {
			var deleteFunc func(string) (int, error)
			if command.IsBatch {
				deleteFunc = func(file string) (int, error) {
					return dispatcher.deleter.Batch(file, command.Deletions)
				}
			} else {
				deleteFunc = func(file string) (int, error) {
					return dispatcher.deleter.Single(file, command.Pattern, command.NegateContent)
				}
			}

			total, err := dispatcher.transactioner.Delete(files, deleteFunc, &stats)
			if err != nil {
				return err
			}
			dispatcher.utils.PrintDeleteMessage(total)
			dispatcher.utils.PrintStats(stats)
			return nil
		}

		errorCollection := models.NewErrorCollection()
		total := 0
		for _, file := range files {
			var count int
			var err error

			if command.IsBatch {
				count, err = dispatcher.deleter.Batch(file, command.Deletions)
			} else {
				count, err = dispatcher.deleter.Single(file, command.Pattern, command.NegateContent)
			}

			if err != nil {
				errorCollection.Add(err)
				stats.Skipped++
				continue
			}

			total += count
			stats.Processed++
		}

		if errorCollection.HasErrors() {
			dispatcher.utils.PrintDeleteMessage(total)
			dispatcher.utils.PrintStats(stats)
			return errorCollection
		}

		dispatcher.utils.PrintDeleteMessage(total)
		dispatcher.utils.PrintStats(stats)
		return nil
	}

	return fmt.Errorf("unhandled command action: %v", command.Action)
}
