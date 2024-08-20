package preprocessor

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	extfiles "github.com/pennsieve/processor-pre-external-files/models"
	"github.com/pennsieve/processor-pre-ttl-sync/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
func TestRun(t *testing.T) {
	datasetId := newDatasetID()

	integrationID := uuid.NewString()
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	sessionToken := uuid.NewString()
	mockServer := newMockServer(t, integrationID, datasetId)
	defer mockServer.Close()

	ttlSyncPP := NewTTLSyncPreProcessor(integrationID, inputDir, outputDir, sessionToken, mockServer.URL, mockServer.URL)

	require.NoError(t, ttlSyncPP.Run())

	expectedExternalFilesConfigPath := filepath.Join(inputDir, ExternalFilesConfigName)
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
			assert.Equal(t, fmt.Sprintf(TTLEndpointPattern, expectedDatasetUUID, expectedName), actualURLConfig.URL)
			assert.Nil(t, actualURLConfig.Auth)
			assert.Nil(t, actualURLConfig.Query)
		}
	}

}

func newDatasetID() string {
	return newDatasetIDWithUUID(uuid.NewString())
}

func newDatasetIDWithUUID(datasetUUID string) string {
	return fmt.Sprintf("%s%s", DatasetNodeIDPrefix, datasetUUID)
}

func newMockServer(t *testing.T, integrationID string, datasetID string) *httptest.Server {
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
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		require.Fail(t, "unexpected call to Pennsieve", "%s %s", request.Method, request.URL)
	})
	return httptest.NewServer(mux)
}
