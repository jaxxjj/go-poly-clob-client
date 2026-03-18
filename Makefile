.PHONY: test lint lint-fix build vet fmt

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

lint-fix:
	@echo "--> Running linter auto fix"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found."; \
		exit 1; \
	fi
	@golangci-lint run --fix

build:
	go build ./...

vet:
	go vet ./...

fmt:
	gofmt -w .
	goimports -w .
