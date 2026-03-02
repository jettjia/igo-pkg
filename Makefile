PKG := "github.com/jettjia/go-pkg"
PKG_LIST := $(shell go list ${PKG}/...)

init:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/jstemmer/go-junit-report/v2@latest
	@go install github.com/matm/gocov-html/cmd/gocov-html@latest

.PHONY: tidy
tidy:
	$(eval files=$(shell find . -name go.mod))
	@set -e; \
	for file in ${files}; do \
		goModPath=$$(dirname $$file); \
		cd $$goModPath; \
		go mod tidy; \
		cd -; \
	done

.PHONY: fmt
fmt:
	@go fmt ${PKG_LIST}

.PHONY: generate
vet: ## generate the files
	@go generate ./...

.PHONY: test
test:
	@go test -cover ./...

.PHONY: race
race: ## Run tests with data race detector
	@go test -race ${PKG_LIST}

.PHONY: test-coverage
test-coverage:  ## Generate a single test report file in HTML format
	@go test ./... -v -coverprofile=report/cover 2>&1 | go-junit-report > report/ut_report.xml
	@gocov convert report/cover | gocov-html > report/coverage.html

.PHONY: golangci
golangci:
	@golangci-lint run --config .golangci.yml