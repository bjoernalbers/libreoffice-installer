PROJECT_NAME     := libreoffice-installer
IDENTIFIER       := de.bjoernalbers.$(PROJECT_NAME)
# Regex to capture Semantic Version string taken from: https://semver.org
VERSION          := $(shell git describe --tags | grep -Eo '^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$$' | tr -d v)
BUILD_DIR        := $(shell mktemp -d)
SCRIPTS_DIR      := $(shell mktemp -d)
EXECUTABLE       := $(BUILD_DIR)/$(PROJECT_NAME)
COMPONENT_PKG    := $(BUILD_DIR)/$(PROJECT_NAME).pkg
DISTRIBUTION_PKG := $(PROJECT_NAME).pkg
TEST_VOLUME      := testvolume

ifndef APPLE_INSTALLER_IDENTITY
$(error APPLE_INSTALLER_IDENTITY is not set)
endif

ifndef APPLE_APPLICATION_IDENTITY
$(error APPLE_APPLICATION_IDENTITY is not set)
endif

.PHONY: $(DISTRIBUTION_PKG) check install clean

$(DISTRIBUTION_PKG):
ifndef VERSION
	$(error No Semantic Version found in git tag)
endif
	$(foreach arch,arm64 amd64, GOARCH="$(arch)" go build -o "$(EXECUTABLE)-$(arch)"; )
	lipo "$(EXECUTABLE)"-* -create -output "$(EXECUTABLE)"
	codesign --sign "$(APPLE_APPLICATION_IDENTITY)" --options=runtime "$(EXECUTABLE)"
	cp "$(EXECUTABLE)" "$(SCRIPTS_DIR)/postinstall"
	pkgbuild \
		--identifier "$(IDENTIFIER)" \
		--version "$(VERSION)" \
		--scripts "$(SCRIPTS_DIR)" \
		--sign "$(APPLE_INSTALLER_IDENTITY)" \
		--quiet \
		--nopayload \
		"$(COMPONENT_PKG)"
	productbuild \
		--package "$(COMPONENT_PKG)" \
		--sign "$(APPLE_INSTALLER_IDENTITY)" \
		--quiet \
		"$@"
	rm -rf "$(BUILD_DIR)" "$(SCRIPTS_DIR)"
	xcrun notarytool submit "$@" \
		--keychain-profile "default" \
		--wait
	xcrun stapler staple "$@"
	spctl --assess --type install "$@"

check: $(DISTRIBUTION_PKG)
	hdiutil create -size 1g testvolume.dmg
	hdiutil attach testvolume.dmg -nobrowse -mountpoint "$(TEST_VOLUME)"
	mkdir -p "$(TEST_VOLUME)/Applications"
	sudo installer -pkg "$(DISTRIBUTION_PKG)" -target "$(TEST_VOLUME)"
	"$(TEST_VOLUME)/Applications/LibreOffice.app/Contents/MacOS/soffice" --version
	hdiutil detach "$(TEST_VOLUME)"
	rm testvolume.dmg

install: $(DISTRIBUTION_PKG)
	sudo installer -pkg "$(DISTRIBUTION_PKG)" -target /

clean:
	rm -rf $(BUILD_DIR)
	-hdiutil detach "$(TEST_VOLUME)" 2>/dev/null
	rm -f testvolume.dmg
