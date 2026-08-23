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
	searcher := commands.NewSearcher(parallelizer, sorter, utils)
	counter := commands.NewCounter(parallelizer, searcher)
	updater := commands.NewUpdater(processor, utils)
	deleter := commands.NewDeleter(processor, utils)

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
