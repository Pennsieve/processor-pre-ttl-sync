# TTL Sync Pre-Processor

Given a dataset node ID, produces the external URLs for the latest TTL files in the format expected
by [processor-pre-external-files](https://github.com/Pennsieve/processor-pre-external-files)

The file is written to `$INPUT_DIR/external-files.json`

To build:

`docker build -t pennsieve/ttl-sync-pre-processor .`

On arm64 architectures:

`docker build -f Dockerfile_arm64 -t pennsieve/ttl-sync-pre-processor .`

To run tests:

` go test ./...`

To run integration test:

1. Given a dataset you want to test with, create an integration for the dataset and this pre-processor. Get the
   integration id
2. Copy `dev.env.example` to `dev.env`
3. In dev.env update SESSION_TOKEN with a valid token and INTEGRATION_ID with the id from the first step.
4. Run `./run-integration-test.sh dev.env`

If `ENVIRONMENT=dev` is in `dev.env`, then the processor will try to fetch the curation-export files from a test server
and you will have to make sure that there are files there to find.

If `ENVIRONMENT=prod` is in `dev.env', the processor will use the real SPARC endpoint to find the curation-export files.