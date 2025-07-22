.PHONY: setup lint test

setup: bin/golangci-lint
	go mod download

bin:
	mkdir bin

bin/golangci-lint: bin
	GOBIN=$(PWD)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0

lint: bin/golangci-lint
	bin/golangci-lint fmt
	go vet ./...
	go mod tidy
	bin/golangci-lint -c .golangci.yml run ./...

test:
	go test -timeout=10s -race -cover ./...
