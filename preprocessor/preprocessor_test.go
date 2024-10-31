package preprocessor

import (
	crypto "crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	extfiles "github.com/pennsieve/processor-pre-external-files/client/models"
	"github.com/pennsieve/processor-pre-ttl-sync/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractDatasetUUID(t *testing.T) {
	t.Run("correct format", func(t *testing.T) {
		expectedDatasetUUID := uuid.NewString()
		datasetUUID, err := ExtractDatasetUUID(newDatasetIDWithUUID(expectedDatasetUUID))
		require.NoError(t, err)
		assert.Equal(t, expectedDatasetUUID, datasetUUID)
	})

	t.Run("incorrect format", func(t *testing.T) {
		wrongFormatID := uuid.NewString()
		_, err := ExtractDatasetUUID(wrongFormatID)
		assert.ErrorContainsf(t, err, DatasetNodeIDPrefix, "error message missing Node ID Prefix")
		assert.ErrorContainsf(t, err, wrongFormatID, "error message missing passed in datasetID")

	})
}

func TestNewTTLSyncPreProcessor(t *testing.T) {
	integrationID := uuid.NewString()
	inputDirectory := uuid.NewString()
	outputDirectory := uuid.NewString()
	sessionToken := uuid.NewString()
	apiHost := uuid.NewString()
	api2Host := uuid.NewString()
	ttlHost := uuid.NewString()

	p := NewTTLSyncPreProcessor(integrationID, inputDirectory, outputDirectory, sessionToken, apiHost, api2Host, ttlHost)

	assert.Equal(t, integrationID, p.IntegrationID)
	assert.Equal(t, inputDirectory, p.InputDirectory)
	assert.Equal(t, outputDirectory, p.OutputDirectory)
	assert.Equal(t, sessionToken, p.Pennsieve.Token)
	assert.Equal(t, apiHost, p.Pennsieve.APIHost)
	assert.Equal(t, api2Host, p.Pennsieve.API2Host)
	assert.Equal(t, ttlHost+TTLEndpointPattern, p.TTLURLPattern)
	assert.Equal(t, inputDirectory, p.ExternalFileProcessor.InputDirectory)
	assert.Equal(t, filepath.Join(inputDirectory, ExternalFilesConfigName), p.ExternalFileProcessor.ConfigFile)
}

func TestRun(t *testing.T) {
	datasetUUID := uuid.NewString()
	datasetId := newDatasetIDWithUUID(datasetUUID)

	integrationID := uuid.NewString()
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	sessionToken := uuid.NewString()
	expectedTTLFiles := newExpectedTTLFiles(t, datasetUUID, inputDir)

	// empty graph schema. We're not testing the metadata processor, just that we are configuring
	// it with the correct parameters
	expectedFiles := append(expectedTTLFiles, ExpectedFile{
		filePath: filepath.Join(inputDir, "metadata", "schema", "graphSchema.json"),
		urlPath:  fmt.Sprintf("/models/datasets/%s/concepts/schema/graph", datasetId),
		content:  []byte("[]"),
	},
		ExpectedFile{
			filePath: filepath.Join(inputDir, "metadata", "schema", "relationships.json"),
			urlPath:  fmt.Sprintf("/models/datasets/%s/relationships", datasetId),
			content:  []byte("[]"),
		})

	mockServer := newMockServer(t, integrationID, datasetId, expectedFiles)
	defer mockServer.Close()

	ttlSyncPP := NewTTLSyncPreProcessor(integrationID, inputDir, outputDir, sessionToken, mockServer.URL, mockServer.URL, mockServer.URL)
	require.NoError(t, ttlSyncPP.Run())

	expectedExternalFilesConfigPath := ttlSyncPP.ExternalFileProcessor.ConfigFile
	if assert.FileExists(t, expectedExternalFilesConfigPath) {
		actual, err := os.Open(expectedExternalFilesConfigPath)
		require.NoError(t, err)
		var actualConfig extfiles.ExternalFileParams
		require.NoError(t, json.NewDecoder(actual).Decode(&actualConfig))

		expectedDatasetUUID, err := ExtractDatasetUUID(datasetId)
		require.NoError(t, err)

		assert.Len(t, actualConfig, len(TTLFileNames))
		for i := range TTLFileNames {
			expectedName := TTLFileNames[i]
			actualURLConfig := actualConfig[i]
			assert.Equal(t, expectedName, actualURLConfig.Name)
			assert.Equal(t, fmt.Sprintf(ttlSyncPP.TTLURLPattern, expectedDatasetUUID, expectedName), actualURLConfig.URL)
			assert.Nil(t, actualURLConfig.Auth)
			assert.Nil(t, actualURLConfig.Query)
		}
	}
	for _, expectedFile := range expectedFiles {
		if assert.FileExists(t, expectedFile.filePath) {
			actualContent, err := os.ReadFile(expectedFile.filePath)
			if assert.NoError(t, err) {
				assert.Equal(t, expectedFile.content, actualContent)
			}

		}
	}

}

func newDatasetID() string {
	return newDatasetIDWithUUID(uuid.NewString())
}

func newDatasetIDWithUUID(datasetUUID string) string {
	return fmt.Sprintf("%s%s", DatasetNodeIDPrefix, datasetUUID)
}

type ExpectedFile struct {
	filePath string
	urlPath  string
	content  []byte
}

func newExpectedTTLFiles(t *testing.T, datasetUUID string, inputDirectory string) []ExpectedFile {
	var files []ExpectedFile
	for _, ttlFileName := range TTLFileNames {
		size := rand.Intn(1000) + 1
		bytes := make([]byte, size)
		_, err := crypto.Read(bytes)
		require.NoError(t, err)

		files = append(files, ExpectedFile{
			filePath: filepath.Join(inputDirectory, ttlFileName),
			urlPath:  fmt.Sprintf(TTLEndpointPattern, datasetUUID, ttlFileName),
			content:  bytes,
		})
	}
	return files
}

func newMockServer(t *testing.T, integrationID string, datasetID string, expectedFiles []ExpectedFile) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/integrations/%s", integrationID), func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method, "expected method %s for %s, got %s", http.MethodGet, request.URL, request.Method)
		integration := models.Integration{
			Uuid:          uuid.NewString(),
			ApplicationID: 0,
			DatasetNodeID: datasetID,
		}
		integrationResponse, err := json.Marshal(integration)
		require.NoError(t, err)
		_, err = writer.Write(integrationResponse)
		require.NoError(t, err)
	})
	for i := range expectedFiles {
		// work around Go loop variable gotcha
		expectedFile := expectedFiles[i]
		mux.HandleFunc(expectedFile.urlPath, func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodGet, request.Method, "expected method %s for %s, got %s", http.MethodGet, request.URL, request.Method)
			_, err := writer.Write(expectedFile.content)
			require.NoError(t, err)
		})
	}
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		require.Fail(t, "unexpected call to mockServer", "%s %s", request.Method, request.URL)
	})
	return httptest.NewServer(mux)
}
