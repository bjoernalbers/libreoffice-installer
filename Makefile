BUILD_DIR := build
TARGET_EXEC := libreoffice-installer
SRC := $(shell find . -name '*.go' -or -name go.mod -or -name go.sum)

.PHONY: clean

$(BUILD_DIR)/$(TARGET_EXEC): $(SRC)
	mkdir -p $(BUILD_DIR)
	GOARCH=arm64 go build -o "$@-arm64"
	GOARCH=amd64 go build -o "$@-amd64"
	lipo "$@"-* -create -output "$@"

clean:
	rm -rf $(BUILD_DIR)
