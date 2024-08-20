package preprocessor

import (
	"encoding/json"
	"fmt"
	extfiles "github.com/pennsieve/processor-pre-external-files/models"
	"github.com/pennsieve/processor-pre-ttl-sync/logging"
	"github.com/pennsieve/processor-pre-ttl-sync/pennsieve"
	"github.com/pennsieve/processor-pre-ttl-sync/util"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var logger = logging.PackageLogger("preprocessor")

const DatasetNodeIDPrefix = "N:dataset:"

const TTLEndpointPattern = "https://cassava.ucsd.edu/sparc/datasets/%s/LATEST/%s"

var TTLFileNames = []string{"curation-export.ttl", "curation-export.json"}

const ExternalFilesConfigName = "external-files.json"

type TTLSyncPreProcessor struct {
	IntegrationID   string
	InputDirectory  string
	OutputDirectory string
	Pennsieve       *pennsieve.Session
}

func NewTTLSyncPreProcessor(
	integrationID string,
	inputDirectory string,
	outputDirectory string,
	sessionToken string,
	apiHost string,
	api2Host string) *TTLSyncPreProcessor {
	session := pennsieve.NewSession(sessionToken, apiHost, api2Host)
	return &TTLSyncPreProcessor{
		IntegrationID:   integrationID,
		InputDirectory:  inputDirectory,
		OutputDirectory: outputDirectory,
		Pennsieve:       session,
	}
}

func (m *TTLSyncPreProcessor) Run() error {
	logger.Info("processing integration", slog.String("integrationID", m.IntegrationID))
	integration, err := m.Pennsieve.GetIntegration(m.IntegrationID)
	if err != nil {
		return err
	}
	datasetID := integration.DatasetNodeID
	logger.Info("constructing external file config for dataset", slog.String("datasetID", datasetID))

	datasetUUID, err := ExtractDatasetUUID(datasetID)
	if err != nil {
		return err
	}
	var externalFiles extfiles.ExternalFileParams
	for _, ttlFile := range TTLFileNames {
		ttlFileURL := fmt.Sprintf(TTLEndpointPattern, datasetUUID, ttlFile)
		externalFiles = append(externalFiles, extfiles.ExternalFileParam{
			URL:  ttlFileURL,
			Name: ttlFile,
		})
		logger.Info("added TTL file URL", slog.String("url", ttlFileURL))
	}

	externalFilesConfigPath := filepath.Join(m.InputDirectory, ExternalFilesConfigName)
	externalFilesConfigFile, err := os.Create(externalFilesConfigPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", externalFilesConfigPath, err)
	}
	defer util.CloseFileAndWarn(externalFilesConfigFile)

	if err := json.NewEncoder(externalFilesConfigFile).Encode(externalFiles); err != nil {
		return fmt.Errorf("error encoding external file config to path %s: %w", externalFilesConfigPath, err)
	}
	logger.Info("wrote external files config", slog.String("path", externalFilesConfigPath))

	return nil
}

func ExtractDatasetUUID(datasetID string) (string, error) {
	if datasetUUID, asExpected := strings.CutPrefix(datasetID, DatasetNodeIDPrefix); asExpected {
		return datasetUUID, nil
	} else {
		return "", fmt.Errorf("datasetID %s missing expected prefix: %s", datasetID, DatasetNodeIDPrefix)
	}
}
