.PHONY: build install clean test

# Binary name
BINARY_NAME=githelper

# Build directory
BUILD_DIR=bin

# Go build flags
BUILD_FLAGS=-v

# Installation directory (usually in PATH)
INSTALL_DIR=$(HOME)/.local/bin

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