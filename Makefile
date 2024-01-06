DATE := $(shell date +"%Y-%m-%d")
TIME := $(shell date +"%H%M%S")
LOGS_DIR := logs/$(DATE)/$(TIME)

.PHONY: all

all: clean build

clean:
	rm -rf build

prebuild:
	mkdir -p build
	go work sync

prerun:
	mkdir -p $(LOGS_DIR)

build_game: prebuild
	cd ./src/game && go mod tidy
	go build -C ./src/game -o ../build/game main.go

build_server: prebuild
	cd ./src/server && go mod tidy
	go build -C ./src/erver -o ../build/server main.go

build: build_game build_server

run_game: prerun build_game
	./build/game 2>&1 | tee -a $(LOGS_DIR)/game.log
	@echo

run_server: prerun build_server
	./build/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo

run: prerun build
	( ./build/game 2>&1 | tee -a $(LOGS_DIR)/game.log ) > /dev/null & disown
	./build/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo

run_with_logs: prerun build
	./build/game & disown
	./build/server
	@echo