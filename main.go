package main

import (
	"github.com/pennsieve/processor-pre-ttl-sync/logging"
	"github.com/pennsieve/processor-pre-ttl-sync/preprocessor"
	"log/slog"
	"os"
)

var logger = logging.PackageLogger("main")

func main() {
	m, err := preprocessor.FromEnv()
	if err != nil {
		logger.Error("error creating preprocessor", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("created TTLSyncPreProcessor",
		slog.String("integrationID", m.IntegrationID),
		slog.String("inputDirectory", m.InputDirectory),
		slog.String("outputDirectory", m.OutputDirectory),
		slog.String("apiHost", m.Pennsieve.APIHost),
		slog.String("api2Host", m.Pennsieve.API2Host),
		slog.String("curationExportURLPattern", m.TTLURLPattern),
	)

	if err := m.Run(); err != nil {
		logger.Error("error running preprocessor", slog.Any("error", err))
		os.Exit(1)
	}
}
