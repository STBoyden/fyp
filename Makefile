SHELL := $(shell which bash)
DATE := $(shell date +"%Y-%m-%d")
TIME := $(shell date +"%H%M%S")
ROOT := $(shell pwd)
LOGS_DIR := $(ROOT)/logs/$(DATE)/$(TIME)
BUILD_DIR := $(ROOT)/build

# server environment variables
export TCP_PORT := 8000
export UDP_PORT := 8001

# game environment variables
export SERVER_ADDRESS := 127.0.0.1
export SERVER_TCP_PORT := $(TCP_PORT)
export SERVER_UDP_PORT := $(UDP_PORT)

.PHONY: all

pre:
	@./scripts/pre

fmt:
	find . -iname *.go -exec gofumpt -w -extra {} \;

lint:
	golangci-lint run

all: pre clean build

clean:
	rm -rf $(BUILD_DIR)

prebuild: pre
	mkdir -p $(BUILD_DIR)
	go mod tidy

prerun: pre
	mkdir -p $(LOGS_DIR)
	go mod tidy

build_game: prebuild
	go build -race -o $(BUILD_DIR)/game src/cmd/game/main.go

build_server: prebuild
	go build -race -o $(BUILD_DIR)/server src/cmd/server/main.go

build: build_game build_server

run_game: build_game prerun
	$(BUILD_DIR)/game 2>&1 | tee -a $(LOGS_DIR)/game.log
	@echo

run_server: build_server prerun
	$(BUILD_DIR)/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo

run: build prerun
	($(BUILD_DIR)/server 2>&1 | tee -a $(LOGS_DIR)/server.log) > /dev/null & disown
	($(BUILD_DIR)/game 2>&1 | tee -a $(LOGS_DIR)/game.log) > /dev/null
	@echo

run_with_logs: build prerun
	$(BUILD_DIR)/server & disown
	$(BUILD_DIR)/game
	@echo
