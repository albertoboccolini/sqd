package mock

import (
	"github.com/overthinkinglabs/sqd/src/models"
	"github.com/overthinkinglabs/sqd/src/services"
	"github.com/overthinkinglabs/sqd/src/services/dry_mode"
)

func NewDryModeRunner() *dry_mode.Runner {
	defaultConfig := models.NewDefaultConfig()
	utils := services.NewUtils(defaultConfig)
	dryModeFileReader := dry_mode.NewFileReader(utils)
	dryModeChangeProcessor := dry_mode.NewChangeProcessor(dryModeFileReader, utils)
	return dry_mode.NewRunner(dryModeChangeProcessor, utils)
}
