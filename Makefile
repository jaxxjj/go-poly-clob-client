.PHONY: test lint build vet

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

build:
	go build ./...

vet:
	go vet ./...

fmt:
	gofmt -w .
	goimports -w .
