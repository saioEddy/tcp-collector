.PHONY: build build-linux clean run

# 二进制文件名
BINARY_NAME=tcp-collector
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

clean:
	rm -rf $(BUILD_DIR)
	go clean

run:
	go run cmd/main.go -config config.yaml

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...
