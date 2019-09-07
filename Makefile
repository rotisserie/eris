PKGS = $(shell go list ./... | grep -v vendor)
SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

check: fmt lint test

fmt:
	@echo Running gofmt
	@go fmt $(SRCDIRS)

lint:
	@echo Running golangci linter
	@go get github.com/golangci/golangci-lint
	@golangci-lint run ./... --deadline 5m

test:
	@echo Running tests
	@go test -race $(PKGS)
