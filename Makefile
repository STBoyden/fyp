DATE := $(shell date +"%Y-%m-%d")
TIME := $(shell date +"%H%M%S")
ROOT := $(shell pwd)
LOGS_DIR := $(ROOT)/logs/$(DATE)/$(TIME)
BUILD_DIR := $(ROOT)/build

.PHONY: all

all: clean build

clean:
	rm -rf $(BUILD_DIR)

prebuild:
	mkdir -p $(BUILD_DIR)
	go work sync

prerun:
	mkdir -p $(LOGS_DIR)

build_game: prebuild
	cd ./src/game && go mod tidy
	go build -C ./src/game -o $(BUILD_DIR)/game main.go

build_server: prebuild
	cd ./src/server && go mod tidy
	go build -C ./src/server -o $(BUILD_DIR)/server main.go

build: build_game build_server

run_game: prerun build_game
	$(BUILD_DIR)/game 2>&1 | tee -a $(LOGS_DIR)/game.log
	@echo

run_server: prerun build_server
	$(BUILD_DIR)/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo

run: prerun build
	( $(BUILD_DIR)/server 2>&1 | tee -a $(LOGS_DIR)/server.log ) > /dev/null & disown
	($(BUILD_DIR)/game 2>&1 | tee -a $(LOGS_DIR)/game.log) > /dev/null
	@echo

run_with_logs: prerun build
	$(BUILD_DIR)/server & disown
	$(BUILD_DIR)/game
	@echo