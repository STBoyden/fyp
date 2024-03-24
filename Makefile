SHELL := $(shell which bash)
DATE := $(shell date +"%Y-%m-%d")
TIME := $(shell date +"%H%M%S")
ROOT := $(shell pwd)
LOGS_DIR := $(ROOT)/logs/$(DATE)/$(TIME)
BUILD_DIR := $(ROOT)/build

.PHONY: all

install_formatter:
	go install mvdan.cc/gofumpt@latest

install_linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.56.2

install_godoc:
	go install golang.org/x/tools/cmd/godoc@latest

install_goenums:
	go install github.com/zarldev/goenums@latest

install_tools: install_formatter install_linter install_godoc install_goenums

check: fmt lint

doc:
	@echo "Documentation hosted on http://127.0.0.1:3000/pkg/fyp/src/"
	@echo
	@godoc -index -http=:3000

pre:
	@./scripts/pre

fmt:
	find . -iname *.go -exec gofumpt -w -extra {} \;

lint:
	golangci-lint run

all: pre clean build

clean:
	rm -rf $(BUILD_DIR)

generate_resources:
	go generate resources/resources_gen.go

generate_enums:
	go generate fyp/src/common/ctypes/state

generate: generate_resources generate_enums

gen: generate

prebuild: generate_resources pre
	mkdir -p $(BUILD_DIR)
	go mod tidy

prerun: pre
	mkdir -p $(LOGS_DIR)
	go mod tidy

build_game: prebuild
	go build -race -o $(BUILD_DIR)/game src/cmd/client/main.go

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
	@pkill server
	@echo

run_with_logs: build prerun
	$(BUILD_DIR)/server & disown
	$(BUILD_DIR)/game
	@pkill server
	@echo
