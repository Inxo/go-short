.PHONY: all clean

APP_NAME := my-url-shortener
SRC_DIR := src
BUILD_DIR := build
GOFILES := $(shell find $(SRC_DIR) -name '*.go')

all: $(BUILD_DIR)/$(APP_NAME)

$(BUILD_DIR)/$(APP_NAME): $(GOFILES)
	@mkdir -p $(BUILD_DIR)
	go build -o $@ $(SRC_DIR)/main.go

clean:
	rm -rf $(BUILD_DIR)
