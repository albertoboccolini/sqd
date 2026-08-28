package mock

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/commands"
	"github.com/overthinkinglabs/sqd/src/services/files"
)

func NewDispatcher() *commands.Dispatcher {
	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)
	processor := files.NewProcessor(utils)

	parallelizer := files.NewParallelizer(utils)
	dryModeRunner := NewDryModeRunner()
	transactioner := commands.NewTransactioner(utils)
	sorter := commands.NewSorter()
	lineReader := commands.NewLineReader()
	searcher := commands.NewSearcher(parallelizer, sorter, utils, lineReader)
	counter := commands.NewCounter(parallelizer, searcher, lineReader)
	updater := commands.NewUpdater(processor, utils, lineReader)
	deleter := commands.NewDeleter(processor, utils, lineReader)

	return commands.NewDispatcher(
		searcher,
		counter,
		updater,
		deleter,
		transactioner,
		dryModeRunner,
		utils,
		parallelizer,
	)
}
