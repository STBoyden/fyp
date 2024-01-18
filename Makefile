DATE := $(shell date +"%Y-%m-%d")
TIME := $(shell date +"%H%M%S")
ROOT := $(shell pwd)
LOGS_DIR := $(ROOT)/logs/$(DATE)/$(TIME)
BUILD_DIR := $(ROOT)/build

# server environment variables
TCP_PORT := 8080
UDP_PORT := 8081
SERVER_ENV := TCP_PORT=${TCP_PORT} UDP_PORT=${UDP_PORT}

# game environment variables
SERVER_ADDRESS := 127.0.0.1
SERVER_TCP_PORT := $(TCP_PORT)
SERVER_UDP_PORT := $(UDP_PORT)
CLIENT_ENV := SERVER_ADDRESS=${SERVER_ADDRESS} SERVER_TCP_PORT=${SERVER_TCP_PORT} SERVER_UDP_PORT=${SERVER_UDP_PORT}

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
	$(CLIENT_ENV) $(BUILD_DIR)/game 2>&1 | tee -a $(LOGS_DIR)/game.log
	@echo

run_server: prerun build_server
	$(SERVER_ENV) $(BUILD_DIR)/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo

run: prerun build
	($(SERVER_ENV) $(BUILD_DIR)/server 2>&1 | tee -a $(LOGS_DIR)/server.log) > /dev/null & disown
	($(CLIENT_ENV) $(BUILD_DIR)/game 2>&1 | tee -a $(LOGS_DIR)/game.log) > /dev/null
	@echo

run_with_logs: prerun build
	$(SERVER_ENV) $(BUILD_DIR)/server & disown
	$(CLIENT_ENV) $(BUILD_DIR)/game
	@echo