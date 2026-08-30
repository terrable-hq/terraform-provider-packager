.PHONY: build fmt test test-integration

build:
	go build -o terraform-provider-packager .

fmt:
	go fmt ./...

test:
	go test ./...

test-integration:
	PACKAGER_INTEGRATION=1 go test ./internal/bundle -run TestRolldownBuildsTypeScriptLambdaArtifact -v
