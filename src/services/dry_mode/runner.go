package dry_mode

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type Runner struct {
	changeProcessor *ChangeProcessor
	utils           *services.Utils
}

func NewRunner(changeProcessor *ChangeProcessor, utils *services.Utils) *Runner {
	return &Runner{
		changeProcessor: changeProcessor,
		utils:           utils,
	}
}

func (runner *Runner) printSummary(action models.TokenType, totalChanges int) {
	switch action {
	case models.UPDATE:
		runner.utils.PrintUpdateMessage(totalChanges)
	case models.DELETE:
		runner.utils.PrintDeleteMessage(totalChanges)
	}
}

func (runner *Runner) Validate(command models.Command, files []string, stats *models.ExecutionStats, useTransaction bool, showDetailedOutputInDryMode bool) error {
	totalChanges := 0
	errorCollection := models.NewErrorCollection()

	if showDetailedOutputInDryMode {
		runner.changeProcessor = runner.changeProcessor.WithPrinting()
	}

	for _, file := range files {
		changeCount, err := runner.changeProcessor.ProcessCommand(file, command, stats)
		if err != nil {
			if useTransaction {
				errorCollection.Add(displayable_errors.NewTransactionFailedError(err.Error()))
				return errorCollection
			}

			errorCollection.Add(err)
			continue
		}

		totalChanges += changeCount
		stats.Processed++
	}

	runner.printSummary(command.Action, totalChanges)
	runner.utils.PrintStats(*stats)

	if errorCollection.HasErrors() {
		return errorCollection
	}

	return nil
}
