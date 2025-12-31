# AWS TUI Makefile

# Variables
BINARY_NAME=aws-tui
BUILD_DIR=bin
MAIN_PATH=cmd/aws-tui/main.go
GO_FILES=$(shell find . -name '*.go')

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test run tidy help

all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

## test: Run unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## run: Build and run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

## tidy: Clean up go.mod and go.sum
tidy:
	@echo "Tidying up modules..."
	$(GOMOD) tidy

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' | column -t -s ':'
