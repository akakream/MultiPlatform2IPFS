.PHONY: build run test vendor

BUILD_NAME = multiplatform2ipfs
BUILD_DIR = $(PWD)/bin

build:
	@go build -o $(BUILD_DIR)/$(BUILD_NAME)

clean.bin:
	rm -rf $(BUILD_DIR)/*

clean.export:
	rm -rf ./export/*

run:
	@$(BUILD_DIR)/$(BUILD_NAME) server --port=3002

test:
	go test -v ./... -count=1

vendor:
	@go mod vendor

lint:
	golangci-lint run -v ./...
