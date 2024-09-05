# Define default target
.PHONY: all
all: build

VERSION_FILE := VERSION
BINARY := go-grid
BUILD_DIR := build
BUILD_PATH := $(BUILD_DIR)/$(BINARY)

VERSION := $(shell cat $(VERSION_FILE))

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_PATH) ./cmd/main.go

bump:
	@echo "Current version: $(VERSION)"
	@read -p "Enter new version: " NEW_VERSION; \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	echo "Version bumped to $$NEW_VERSION"

tag:
	@read -p "Enter tag message: " TAG_MSG; \
	git commit -am "Bump version to $(shell cat $(VERSION_FILE))"; \
	git tag -a $(shell cat $(VERSION_FILE)) -m "$$TAG_MSG"; \
	git push origin $(shell cat $(VERSION_FILE)); \
	echo "Tagged with $(shell cat $(VERSION_FILE))"

clean:
	rm -rf $(BUILD_DIR)
