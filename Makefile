BUILD_DIR := build
TARGET_EXEC := libreoffice-installer

.PHONY: clean

$(BUILD_DIR)/$(TARGET_EXEC):
	mkdir -p $(BUILD_DIR)
	GOARCH=arm64 go build -o "$@_arm64"
	GOARCH=amd64 go build -o "$@_amd64"
	lipo "$@_arm64" "$@_amd64" -create -output "$@"

clean:
	rm -r $(BUILD_DIR)
