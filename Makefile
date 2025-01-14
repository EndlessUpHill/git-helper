.PHONY: build install clean test

# Binary name
BINARY_NAME=githelper

# Build directory
BUILD_DIR=bin

# Go build flags
BUILD_FLAGS=-v -ldflags "-X github.com/EndlessUphill/git-helper/internal/version.Version=${VERSION} \
                        -X github.com/EndlessUphill/git-helper/internal/version.CommitHash=${COMMIT_HASH} \
                        -X github.com/EndlessUphill/git-helper/internal/version.BuildDate=${BUILD_DATE}"

# Installation directory (usually in PATH)
INSTALL_DIR=$(HOME)/.local/bin

# Add these variables at the top
VERSION=$(shell git describe --tags --always --dirty)
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)
	@echo "Installation complete. Make sure $(INSTALL_DIR) is in your PATH"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test:
	@go test ./... -v 