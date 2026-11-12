
BIN := ./bin

.PHONY: install build test tools lint lint-fix run

install:
	go install

tools:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.12.2

build:
	go build -v -o $(BIN)/atom-util main.go

lint:
	$(BIN)/golangci-lint run
	$(BIN)/golangci-lint fmt -d

lint-fix:
	$(BIN)/golangci-lint run --fix
	$(BIN)/golangci-lint fmt

test:
	go test

run:
	go run .