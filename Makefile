DATE := $(shell date +"%Y-%m-%d")
TIME := $(shell date +"%H%M%S")
LOGS_DIR := logs/$(DATE)/$(TIME)

.PHONY: all

all: clean run

clean:
	rm -rf build

prebuild:
	mkdir -p build
	go work sync

prerun:
	mkdir -p $(LOGS_DIR)

build_game: prebuild
	cd ./game && go mod tidy
	go build -C ./game -o ../build/game main.go

build_server: prebuild
	cd ./server && go mod tidy
	go build -C ./server -o ../build/server main.go

run: prerun build_game build_server
	( ./build/game 2>&1 | tee -a $(LOGS_DIR)/game.log ) > /dev/null & disown
	./build/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo

run_server: prerun build_server
	./build/server 2>&1 | tee -a $(LOGS_DIR)/server.log
	@echo
