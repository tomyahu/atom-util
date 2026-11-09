
BIN := ./bin

.PHONY: build test tools lint lint-fix run

tools:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.12.2

build:
	GOOS=linux GOARCH=amd64 go build -v -o $(BIN)/find-line main.go

lint:
	$(BIN)/golangci-lint run
	$(BIN)/golangci-lint fmt -d

lint-fix:
	$(BIN)/golangci-lint run --fix
	$(BIN)/golangci-lint fmt

test:
	go test

run:
	GOOS=linux GOARCH=amd64 go run .