set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

DATE := if os_family() == "unix" { `date +"%Y-%m-%d"` } else { `get-date -format "yyyy-MM-dd"` }
TIME := if os_family() == "unix" { `date +"%H%M%S"` } else { `get-date -format "HH:mm:ss"` }
ROOT := absolute_path(".")
LD_FLAGS := if os_family() == "unix" { `./scripts/ld-flags.sh` } else { "" }
GO_FLAGS := "-race"
LOGS_DIR := ROOT / "logs" / DATE / TIME
BUILD_DIR := ROOT / "build"

pre_script := if os_family() == "unix" { "./scripts/pre.sh" } else { "" }

all: pre clean build

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
  @{{pre_script}}

fmt:
	find . -iname *.go -exec gofumpt -w -extra {} \;

lint:
	golangci-lint run

clean:
	rm -rf "{{BUILD_DIR}}"

generate_resources:
	go generate resources/resources_gen.go

generate_enums:
	go generate fyp/src/common/ctypes/state

generate_tiles:
	@go mod tidy
	go generate fyp/src/common/ctypes
	go generate fyp/src/common/ctypes/tiles

generate: generate_resources generate_enums generate_tiles

gen: generate

prebuild: generate pre
	mkdir -p "{{BUILD_DIR}}"
	go mod tidy

prerun: pre
	mkdir -p {{LOGS_DIR}}
	go mod tidy

build_game: prebuild
	go build {{GO_FLAGS}} -o {{BUILD_DIR}}/game src/cmd/client/main.go

build_server: prebuild
	go build {{GO_FLAGS}} -o {{BUILD_DIR}}/server src/cmd/server/main.go

build: build_game build_server

run_game: build_game prerun
	{{BUILD_DIR}}/game 2>&1 | tee -a {{LOGS_DIR}}/game.log
	@echo

run_server: build_server prerun
	{{BUILD_DIR}}/server 2>&1 | tee -a {{LOGS_DIR}}/server.log
	@echo

[unix]
run: build prerun
	({{BUILD_DIR}}/server 2>&1 | tee -a {{LOGS_DIR}}/server.log) & disown
	({{BUILD_DIR}}/game 2>&1 | tee -a {{LOGS_DIR}}/game.log)
	@pkill server
	@echo

[windows]
run: build prerun
	({{BUILD_DIR}}/server 2>&1 | tee -a {{LOGS_DIR}}/server.log) & disown
	({{BUILD_DIR}}/game 2>&1 | tee -a {{LOGS_DIR}}/game.log)
	@pkill server
	@echo
