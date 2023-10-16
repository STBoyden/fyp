DATE := $(shell date -u +"%Y-%m-%d")
TIME := $(shell date -u +"%H%M%S")
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
	( ./build/game | tee -a $(LOGS_DIR)/game.tmp.log ) > /dev/null & disown
	./build/server | tee -a $(LOGS_DIR)/server.tmp.log
	@echo

run_server: prerun build_server
	./build/server | tee -a $(LOGS_DIR)/server.tmp.log
	@echo
