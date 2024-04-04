set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

date := trim(if os_family() == "unix" { `date +"%Y-%m-%d"` } else { `get-date -format "yyyy-MM-dd"` })
time := trim(if os_family() == "unix" { `date +"%H%M%S"` } else { `get-date -format "HHmmss"` })
root := absolute_path(".")
ld_flags := if os_family() == "unix" { `./scripts/ld-flags.sh` } else { "" }
go_flags := if os_family() == "unix" { "-race" } else { "" }
release_ld_flags := "-ldflags '-s -w'"
logs_dir := root / "logs" / date / time
build_dir := root / "build"
dist_dir := root / "dist"
pre_script := if os_family() == "unix" { "./scripts/pre.sh" } else { "" }
platform := if os_family() == "unix" { lowercase(`uname`) } else { "windows" }
arch := if arch() == "arm" { "arm64" } else { if arch() == "x86_64" { "amd64" } else { arch() } }
ext := if platform == "windows" { ".exe" } else { "" }

alias pkg := package
alias gen := generate
alias b := build
alias r := run

export GOOS := platform
export GOARCH := arch

all: install_tools generate build

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

[private]
pre:
    @{{ pre_script }}

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
    rm -rf "{{ build_dir }}"
    rm -rf "{{ dist_dir }}"

[windows]
clean:
    if (test-path "{{ build_dir }}") { remove-item -recurse -force "{{ build_dir }}" }
    if (test-path "{{ dist_dir }}") { remove-item -recurse -force "{{ dist_dir }}"}

generate:
    @go mod tidy
    go generate resources/resources_gen.go
    go generate fyp/src/common/ctypes
    go generate fyp/src/common/ctypes/state
    go generate fyp/src/common/ctypes/tiles

[private]
prebuild: generate pre
    mkdir -p "{{ build_dir }}"
    go mod tidy

[private]
prerun: pre
    mkdir -p "{{ logs_dir }}"
    go mod tidy

build_game: clean prebuild
    go build {{ ld_flags }} {{ go_flags }} -o {{ build_dir }}/game{{ ext }} src/cmd/client/main.go

build_server: clean prebuild
    go build {{ ld_flags }} {{ go_flags }} -o {{ build_dir }}/server{{ ext }} src/cmd/server/main.go

build: build_game build_server

run_game: build_game prerun
    {{ build_dir }}/game{{ ext }} 2>&1 | tee -a {{ logs_dir }}/game.log
    @echo

run_server: build_server prerun
    {{ build_dir }}/server{{ ext }} 2>&1 | tee -a {{ logs_dir }}/server.log
    @echo

[unix]
run: build prerun
    ({{ build_dir }}/server 2>&1 | tee -a {{ logs_dir }}/server.log) & disown
    ({{ build_dir }}/game 2>&1 | tee -a {{ logs_dir }}/game.log)
    @pkill server
    @echo

[windows]
run: clean build prerun
    start-process "{{ build_dir }}/server.exe" -RedirectStandardOutput "{{ logs_dir }}/server.out.log" -RedirectStandardError "{{ logs_dir }}/server.err.log"
    start-process -Wait "{{ build_dir }}/game.exe" -RedirectStandardOutput "{{ logs_dir }}/game.out.log" -RedirectStandardError "{{ logs_dir }}/game.err.log"
    kill $(get-process server | select -expand id)
    @echo ""

[private]
prepackage: clean prebuild
    go build {{ release_ld_flags }} -o {{ dist_dir }}/game{{ ext }} src/cmd/client/main.go
    go build {{ release_ld_flags }} -o {{ dist_dir }}/server{{ ext }} src/cmd/server/main.go

[unix]
package: prepackage
    cp -r resources dist/
    cp .env.example dist/.env
    find dist/resources/ -iname *.go -exec rm {} \;
    rm dist/resources/.gitignore
    mv dist/ final_year_project
    zip -9 -r "fyp-{{ platform }}-{{ arch }}.zip" final_year_project
    rm -rf final_year_project/

[windows]
package: prepackage
    copy-item -path "resources" -recurse -exclude "*.go",".gitignore" -destination "dist" 
    copy-item -path ".env.example" -destination "dist/.env"
    rename-item -path "dist" -newname "final_year_project"
    compress-archive "final_year_project" -compressionlevel optimal "fyp-{{ platform }}-{{ arch }}.zip" -force
    remove-item -path "final_year_project" -recurse -force
