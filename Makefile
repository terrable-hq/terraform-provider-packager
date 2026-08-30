.PHONY: build check-fmt ci fmt test test-integration test-race vet

build:
	go build -o terraform-provider-packager .

fmt:
	go fmt ./...

check-fmt:
	test -z "$$(gofmt -l .)"

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

test-integration:
	PACKAGER_INTEGRATION=1 go test ./internal/bundle -run TestRolldownBuildsTypeScriptLambdaArtifact -v

ci: check-fmt test-race vet build test-integration
