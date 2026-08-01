package dry_mode

import (
	"errors"
	"os"
	"strings"

	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
	"github.com/overthinkinglabs/sqd/src/services"
)

type FileReader struct {
	utils *services.Utils
}

func NewFileReader(utils *services.Utils) *FileReader {
	return &FileReader{utils: utils}
}

func (fileReader *FileReader) ValidateAndReadFile(file string) ([]string, error) {
	if !fileReader.utils.IsPathInsideCwd(file) {
		return nil, displayable_errors.NewInvalidPathError(file)
	}

	if !fileReader.utils.CanWriteFile(file) {
		return nil, displayable_errors.NewPermissionDeniedError(file)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, displayable_errors.NewPermissionDeniedError(file)
		}

		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}
