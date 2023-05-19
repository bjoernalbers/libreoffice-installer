PROJECT_NAME := libreoffice-installer
IDENTIFIER := de.bjoernalbers.$(PROJECT_NAME)
IDENTITY_NAME := Developer ID Installer: Bjoern Albers (2M83WXV6U8)
# Regex to capture Semantic Version string taken from: https://semver.org
VERSION := $(shell git describe --tags | grep -Eo '^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$$' | tr -d v)
BUILD_DIR := build
SCRIPTS_DIR := $(BUILD_DIR)/scripts
DIST_DIR := dist
EXECUTABLE := $(BUILD_DIR)/$(PROJECT_NAME)
COMPONENT_PKG := $(BUILD_DIR)/$(PROJECT_NAME).pkg
DISTRIBUTION_PKG := $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION).pkg
TEST_VOLUME := testvolume

.PHONY: check clean

$(DISTRIBUTION_PKG): $(COMPONENT_PKG)
ifndef VERSION
	$(error No Semantic Version found in git tag)
endif
	productbuild \
		--package "$<" \
		--sign "$(IDENTITY_NAME)" \
		--quiet \
		"$@"

$(COMPONENT_PKG): $(EXECUTABLE)
ifndef VERSION
	$(error No Semantic Version found in git tag)
endif
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

$(EXECUTABLE): $(shell find . -name '*.go' -or -name go.mod -or -name go.sum)
	mkdir -p $(BUILD_DIR)
	GOARCH=arm64 go build -o "$@-arm64"
	GOARCH=amd64 go build -o "$@-amd64"
	lipo "$@"-* -create -output "$@"

check: $(DISTRIBUTION_PKG)
	hdiutil create -size 1g testvolume.dmg
	hdiutil attach testvolume.dmg -nobrowse -mountpoint "$(TEST_VOLUME)"
	mkdir -p "$(TEST_VOLUME)/Applications"
	sudo installer -pkg "$(DISTRIBUTION_PKG)" -target "$(TEST_VOLUME)"
	"$(TEST_VOLUME)/Applications/LibreOffice.app/Contents/MacOS/soffice" --version
	hdiutil detach "$(TEST_VOLUME)"
	rm testvolume.dmg

clean:
	rm -rf $(BUILD_DIR)
	-hdiutil detach "$(TEST_VOLUME)" 2>/dev/null
	rm -f testvolume.dmg
