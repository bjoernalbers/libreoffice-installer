PROJECT_NAME := libreoffice-installer
BUILD_DIR := build
SCRIPTS_DIR := $(BUILD_DIR)/scripts
TARGET_EXEC := $(PROJECT_NAME)
SRC := $(shell find . -name '*.go' -or -name go.mod -or -name go.sum)
COMPONENT_PKG := $(PROJECT_NAME).pkg
IDENTIFIER := de.bjoernalbers.$(PROJECT_NAME)
IDENTITY_NAME := Developer ID Installer: Bjoern Albers (2M83WXV6U8)
VERSION := 0.0.1

.PHONY: clean

$(BUILD_DIR)/$(COMPONENT_PKG): $(BUILD_DIR)/$(TARGET_EXEC)
	mkdir -p $(SCRIPTS_DIR)
	cp "$<" "$(SCRIPTS_DIR)/postinstall"
	pkgbuild \
		--nopayload \
		--scripts "$(SCRIPTS_DIR)" \
		--identifier "$(IDENTIFIER)" \
		--version "$(VERSION)" \
		--sign "$(IDENTITY_NAME)" \
		--quiet \
		"$@"

$(BUILD_DIR)/$(TARGET_EXEC): $(SRC)
	mkdir -p $(BUILD_DIR)
	GOARCH=arm64 go build -o "$@-arm64"
	GOARCH=amd64 go build -o "$@-amd64"
	lipo "$@"-* -create -output "$@"

clean:
	rm -rf $(BUILD_DIR)
