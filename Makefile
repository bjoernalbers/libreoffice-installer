BUILD_DIR := build

.PHONY: clean

build:
	mkdir -p $(BUILD_DIR)

clean:
	rm -r $(BUILD_DIR)
