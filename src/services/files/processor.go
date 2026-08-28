package files

import (
	"errors"
	"os"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type Processor struct {
	utils *services.Utils
}

func NewProcessor(utils *services.Utils) *Processor {
	return &Processor{utils: utils}
}

func (processor *Processor) ProcessFile(filename string, transformFunc func([]string) ([]string, int)) (int, error) {
	if !processor.utils.IsPathInsideCwd(filename) {
		return 0, displayable_errors.NewInvalidPathError(filename)
	}
	if !processor.utils.CanWriteFile(filename) {
		return 0, displayable_errors.NewPermissionDeniedError(filename)
	}

	info, err := os.Stat(filename)
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	lines := strings.Split(string(data), "\n")
	newLines, count := transformFunc(lines)

	if count == 0 {
		return 0, nil
	}

	if err := os.WriteFile(filename, []byte(strings.Join(newLines, "\n")), info.Mode().Perm()); err != nil {
		return 0, processor.mapFileError(filename, err)
	}

	return count, nil
}

func (processor *Processor) mapFileError(filename string, err error) error {
	if errors.Is(err, os.ErrPermission) {
		return displayable_errors.NewPermissionDeniedError(filename)
	}

	return err
}
