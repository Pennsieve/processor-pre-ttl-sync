package preprocessor

import (
	"fmt"
	"os"
	"strings"
)

const IntegrationIDKey = "INTEGRATION_ID"
const InputDirectoryKey = "INPUT_DIR"
const OutputDirectoryKey = "OUTPUT_DIR"
const SessionTokenKey = "SESSION_TOKEN"
const PennsieveAPIHostKey = "PENNSIEVE_API_HOST"
const PennsieveAPI2HostKey = "PENNSIEVE_API_HOST2"
const EnvironmentKey = "ENVIRONMENT"
const ProdEnv = "prod"
const DevEnv = "dev"

const ProdTTLHost = "https://cassava.ucsd.edu"
const DevTTLHost = "http://test-sparc-curation-export-source.s3-website-us-east-1.amazonaws.com"

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
	env, err := LookupRequiredEnvVar(EnvironmentKey)
	if err != nil {
		return nil, err
	}
	var ttHost string
	if strings.ToLower(env) == DevEnv {
		ttHost = DevTTLHost
	} else if strings.ToLower(env) == ProdEnv {
		ttHost = ProdTTLHost
	} else {
		return nil, fmt.Errorf("unexpected value for %s; expect either %q or %q (case insensitive): %s",
			EnvironmentKey,
			DevEnv,
			ProdEnv,
			env)
	}
	return NewTTLSyncPreProcessor(integrationID,
		inputDirectory,
		outputDirectory,
		sessionToken,
		apiHost,
		api2Host,
		ttHost), nil
}

func LookupRequiredEnvVar(key string) (string, error) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "", fmt.Errorf("no %s set", key)
	}
	return value, nil
}
