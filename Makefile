GORELEASER ?= goreleaser

.PHONY: build check-fmt ci fmt release-check release-snapshot test test-acceptance test-integration test-race vet

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

test-acceptance:
	TF_ACC=1 go test ./internal/provider -run '^TestAcc' -v -timeout 10m

test-integration:
	PACKAGER_INTEGRATION=1 go test ./internal/bundle -run '^TestEmbeddedBundleExecutesInNode$$' -count=1 -v

ci: check-fmt test-race vet build test-integration test-acceptance

release-check:
	$(GORELEASER) check --config .goreleaser.yml

release-snapshot: release-check
	$(GORELEASER) release --snapshot --clean --skip=sign --config .goreleaser.yml
	PACKAGER_RELEASE_SNAPSHOT=1 go test ./tests/release -count=1 -v
