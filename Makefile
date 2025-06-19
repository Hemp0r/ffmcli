BINARY_NAME=ffmcli
MAIN_PATH=./src/main.go
BUILD_DIR=./build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=cd ./src && $(GOCMD) get -u
GOMOD=cd ./src && $(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: all build clean test deps install uninstall help run check

all: clean deps test build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	cd src && $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: clean deps test
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	cd src && GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	
	# Linux ARM64
	cd src && GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	
	# Windows AMD64
	cd src && GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	
	# Windows ARM64
	cd src && GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe main.go
	
	# macOS (Apple Silicon only)
	cd src && GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	
	@echo "Multi-platform build complete!"
	@ls -la $(BUILD_DIR)/

build-windows: deps
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	cd src && GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME).exe main.go
	@echo "Windows build complete: $(BUILD_DIR)/$(BINARY_NAME).exe"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

deps-update:
	@echo "Updating dependencies..."
	$(GOGET) ./...
	$(GOMOD) tidy

run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) --help

# Check system compatibility
check: build
	@echo "Checking system compatibility..."
	$(BUILD_DIR)/$(BINARY_NAME) check

presets: build
	@echo "Available presets:"
	$(BUILD_DIR)/$(BINARY_NAME) presets

package: build-all
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/packages
	
	# Linux AMD64
	@tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	
	# Linux ARM64
	@tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-arm64
	
	# Windows AMD64
	@zip -j $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	
	# Windows ARM64
	@zip -j $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe
	
	# macOS (Apple Silicon only)
	@tar -czf $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64
	
	@echo "Release packages created in $(BUILD_DIR)/packages/"
	@ls -la $(BUILD_DIR)/packages/

package-windows: build-windows
	@echo "Creating Windows release package..."
	@mkdir -p $(BUILD_DIR)/packages
	@zip -j $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-windows.zip $(BUILD_DIR)/$(BINARY_NAME).exe
	@echo "Windows package created: $(BUILD_DIR)/packages/$(BINARY_NAME)-$(VERSION)-windows.zip"

test:
	@echo "Running tests..."
	@cd src && go test ./...

test-coverage:
	@echo "Running tests with coverage..."
	@cd src && go test -coverprofile=../coverage.out ./...