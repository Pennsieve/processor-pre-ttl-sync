package preprocessor

import (
	"fmt"
	"os"
)

const IntegrationIDKey = "INTEGRATION_ID"
const InputDirectoryKey = "INPUT_DIR"
const OutputDirectoryKey = "OUTPUT_DIR"
const SessionTokenKey = "SESSION_TOKEN"
const PennsieveAPIHostKey = "PENNSIEVE_API_HOST"
const PennsieveAPI2HostKey = "PENNSIEVE_API_HOST2"

const ProdTTLHost = "https://cassava.ucsd.edu"

func FromEnv() (*TTLSyncPreProcessor, error) {
	integrationID, err := LookupRequiredEnvVar(IntegrationIDKey)
	if err != nil {
		return nil, err
	}
	inputDirectory, err := LookupRequiredEnvVar(InputDirectoryKey)
	if err != nil {
		return nil, err
	}
	outputDirectory, err := LookupRequiredEnvVar(OutputDirectoryKey)
	if err != nil {
		return nil, err
	}
	sessionToken, err := LookupRequiredEnvVar(SessionTokenKey)
	if err != nil {
		return nil, err
	}
	apiHost, err := LookupRequiredEnvVar(PennsieveAPIHostKey)
	if err != nil {
		return nil, err
	}
	api2Host, err := LookupRequiredEnvVar(PennsieveAPI2HostKey)
	if err != nil {
		return nil, err
	}
	return NewTTLSyncPreProcessor(integrationID,
		inputDirectory,
		outputDirectory,
		sessionToken,
		apiHost,
		api2Host,
		ProdTTLHost), nil
}

func LookupRequiredEnvVar(key string) (string, error) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "", fmt.Errorf("no %s set", key)
	}
	return value, nil
}
