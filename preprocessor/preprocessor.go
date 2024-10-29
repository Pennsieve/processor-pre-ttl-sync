package preprocessor

import (
	"encoding/json"
	"fmt"
	extfiles "github.com/pennsieve/processor-pre-external-files/client/models"
	extfilesproc "github.com/pennsieve/processor-pre-external-files/service/preprocessor"
	metadataproc "github.com/pennsieve/processor-pre-metadata/service/preprocessor"
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

const TTLEndpointPattern = "/sparc/datasets/%s/LATEST/%s"

var TTLFileNames = []string{"curation-export.json"}

const ExternalFilesConfigName = "external-files.json"

type TTLSyncPreProcessor struct {
	IntegrationID         string
	InputDirectory        string
	OutputDirectory       string
	Pennsieve             *pennsieve.Session
	TTLURLPattern         string
	ExternalFileProcessor *extfilesproc.ExternalFilesPreProcessor
	MetadataProcessor     *metadataproc.MetadataPreProcessor
}

func NewTTLSyncPreProcessor(
	integrationID string,
	inputDirectory string,
	outputDirectory string,
	sessionToken string,
	apiHost string,
	api2Host string,
	ttlHost string) *TTLSyncPreProcessor {
	session := pennsieve.NewSession(sessionToken, apiHost, api2Host)

	metadataPreProcessor := metadataproc.NewMetadataPreProcessor(integrationID, inputDirectory, outputDirectory, sessionToken, apiHost, api2Host, 0)

	externalFilesConfigPath := filepath.Join(inputDirectory, ExternalFilesConfigName)
	extFilePreProcessor := extfilesproc.NewExternalFilesPreProcessor(integrationID, inputDirectory, outputDirectory, externalFilesConfigPath)

	ttlURLPattern := ttlHost + TTLEndpointPattern
	return &TTLSyncPreProcessor{
		IntegrationID:         integrationID,
		InputDirectory:        inputDirectory,
		OutputDirectory:       outputDirectory,
		Pennsieve:             session,
		ExternalFileProcessor: extFilePreProcessor,
		MetadataProcessor:     metadataPreProcessor,
		TTLURLPattern:         ttlURLPattern,
	}
}

func (m *TTLSyncPreProcessor) Run() error {
	logger.Info("processing integration", slog.String("integrationID", m.IntegrationID))
	integration, err := m.Pennsieve.GetIntegration(m.IntegrationID)
	if err != nil {
		return err
	}
	datasetID := integration.DatasetNodeID

	logger.Info("Running TTL sync", slog.String("datasetID", datasetID))

	m.MetadataProcessor = m.MetadataProcessor.WithDatasetID(datasetID)

	if err := m.MetadataProcessor.Run(); err != nil {
		return err
	}

	logger.Info("constructing external file config for dataset", slog.String("datasetID", datasetID))

	datasetUUID, err := ExtractDatasetUUID(datasetID)
	if err != nil {
		return err
	}
	var externalFiles extfiles.ExternalFileParams
	for _, ttlFile := range TTLFileNames {
		ttlFileURL := fmt.Sprintf(m.TTLURLPattern, datasetUUID, ttlFile)
		externalFiles = append(externalFiles, extfiles.ExternalFileParam{
			URL:  ttlFileURL,
			Name: ttlFile,
		})
		logger.Info("added TTL file URL", slog.String("url", ttlFileURL))
	}

	externalFilesConfigPath := m.ExternalFileProcessor.ConfigFile
	externalFilesConfigFile, err := os.Create(externalFilesConfigPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", externalFilesConfigPath, err)
	}
	defer util.CloseFileAndWarn(externalFilesConfigFile)

	if err := json.NewEncoder(externalFilesConfigFile).Encode(externalFiles); err != nil {
		return fmt.Errorf("error encoding external file config to path %s: %w", externalFilesConfigPath, err)
	}
	logger.Info("wrote external files config", slog.String("path", externalFilesConfigPath))
	if err := m.downloadTTLFiles(); err != nil {
		return err
	}
	return nil
}

func (m *TTLSyncPreProcessor) downloadTTLFiles() error {
	logger.Info("downloading TTL files")
	if err := m.ExternalFileProcessor.Run(); err != nil {
		return err
	}
	logger.Info("downloaded TTL files")
	return nil
}

func ExtractDatasetUUID(datasetID string) (string, error) {
	if datasetUUID, asExpected := strings.CutPrefix(datasetID, DatasetNodeIDPrefix); asExpected {
		return datasetUUID, nil
	} else {
		return "", fmt.Errorf("datasetID %s missing expected prefix: %s", datasetID, DatasetNodeIDPrefix)
	}
}
