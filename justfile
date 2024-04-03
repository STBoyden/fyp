set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

DATE := trim(if os_family() == "unix" { `date +"%Y-%m-%d"` } else { `get-date -format "yyyy-MM-dd"` })
TIME := trim(if os_family() == "unix" { `date +"%H%M%S"` } else { `get-date -format "HHmmss"` })
ROOT := absolute_path(".")
LD_FLAGS := if os_family() == "unix" { `./scripts/ld-flags.sh` } else { "" }
GO_FLAGS := if os_family() == "unix" { "-race" } else { "" }
LOGS_DIR := ROOT / "logs" / DATE / TIME
BUILD_DIR := ROOT / "build"
EXT := if os_family() == "windows" { ".exe" } else { "" }

pre_script := if os_family() == "unix" { "./scripts/pre.sh" } else { "" }

all: pre clean build

install_formatter:
	go install mvdan.cc/gofumpt@latest

[unix]
install_linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin v1.56.2

[windows]
install_linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2

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

[unix]
fmt:
	find . -iname *.go -exec gofumpt -w -extra {} \;

[windows]
fmt:
	get-childitem -Filter *.go -Recurse | % { & gofumpt -w -extra $_.FullName }

lint:
	golangci-lint run

[unix]
clean:
	rm -rf "{{BUILD_DIR}}"

[windows]
clean:
	if (test-path "{{BUILD_DIR}}") { remove-item -recurse -force "{{BUILD_DIR}}" }

generate:
	@go mod tidy
	go generate resources/resources_gen.go
	go generate fyp/src/common/ctypes
	go generate fyp/src/common/ctypes/state
	go generate fyp/src/common/ctypes/tiles

gen: generate

prebuild: generate pre
	mkdir -p "{{BUILD_DIR}}"
	go mod tidy

prerun: pre
	mkdir -p "{{LOGS_DIR}}"
	go mod tidy

build_game: clean prebuild
	go build {{LD_FLAGS}} {{GO_FLAGS}} -o {{BUILD_DIR}}/game{{EXT}} src/cmd/client/main.go

build_server: clean prebuild
	go build {{LD_FLAGS}} {{GO_FLAGS}} -o {{BUILD_DIR}}/server{{EXT}} src/cmd/server/main.go

build: build_game build_server

run_game: build_game prerun
	{{BUILD_DIR}}/game{{EXT}} 2>&1 | tee -a {{LOGS_DIR}}/game.log
	@echo

run_server: build_server prerun
	{{BUILD_DIR}}/server{{EXT}} 2>&1 | tee -a {{LOGS_DIR}}/server.log
	@echo

[unix]
run: build prerun
	({{BUILD_DIR}}/server 2>&1 | tee -a {{LOGS_DIR}}/server.log) & disown
	({{BUILD_DIR}}/game 2>&1 | tee -a {{LOGS_DIR}}/game.log)
	@pkill server
	@echo

[windows]
run: clean build prerun
	start-process "{{BUILD_DIR}}/server.exe" -RedirectStandardOutput "{{LOGS_DIR}}/server.out.log" -RedirectStandardError "{{LOGS_DIR}}/server.err.log"
	start-process -Wait "{{BUILD_DIR}}/game.exe" -RedirectStandardOutput "{{LOGS_DIR}}/game.out.log" -RedirectStandardError "{{LOGS_DIR}}/game.err.log"
	kill $(get-process server | select -expand id)
	@echo ""
