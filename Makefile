
CURRENT_PATH = $(shell pwd)
GO = GO111MODULE=on go
TEST_PKGS := $(shell $(GO) list ./... | grep -v 'mock_*' | grep -v 'proto')

test:
	go generate ./...
	$(GO) test ${TEST_PKGS} -race -count=1

## make linter: Run golanci-lint
linter:
	golangci-lint run -E goimports --skip-dirs-use-default --skip-dirs mock_client -D staticcheck
