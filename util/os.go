package util

import (
	"os"
)

func CloseFileAndWarn(file *os.File) {
	if err := file.Close(); err != nil {
		logger.Warn("error closing file %s: %w", file.Name(), err)
	}
}
