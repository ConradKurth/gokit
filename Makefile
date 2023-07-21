.PHONY: generate

help:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

GOLANGCI_VERSION = 1.46.2
FLINT_VERSION = 0.1.1

.bin/golangci-lint: .bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} .bin/golangci-lint
.bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p .bin
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./.bin -- v${GOLANGCI_VERSION}
	@mv .bin/golangci-lint $@

.bin/flint: .bin/flint-${FLINT_VERSION}
	@ln -sf flint-${FLINT_VERSION} .bin/flint
.bin/flint-${FLINT_VERSION}:
	@mkdir -p .bin
	GOBIN=$(PWD)/.bin go install github.com/fraugster/flint@v${FLINT_VERSION}
	@mv .bin/flint $@
	go mod tidy # revert changes from go install

lint: .bin/golangci-lint .bin/flint ## run linter checks
	.bin/flint ./...
	.bin/golangci-lint run

test: ## run unit tests
	go test ./...
