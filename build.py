#!/usr/bin/env python3

import os
import platform
import shutil
import signal
import subprocess
import sys
from datetime import date, datetime

if __name__ != "__main__":
    print("This script needs to be ran directly")
    exit(1)

DATE = date.today().strftime("%Y-%m-%d")
TIME = datetime.now().strftime("%H%M%S")
ROOT = os.path.dirname(os.path.realpath(__file__))
LOGS_DIR = os.path.normpath(f"{ROOT}/logs/{DATE}/{TIME}")
BUILD_DIR = os.path.normpath(f"{ROOT}/build")

SRC_DIR = os.path.normpath(f"{ROOT}/src")
GAME_SRC = os.path.normpath(f"{SRC_DIR}/cmd/client")
SERVER_SRC = os.path.normpath(f"{SRC_DIR}/cmd/server")

# if we're on wsl2, then we have to manually make sure that the GOOS environment variable is set to Linux
if platform.uname().release.endswith("microsoft-standard-WSL2"):
    os.environ["GOOS"] = "linux"

ext = ".exe" if os.name == "nt" else ""


def clean():
    print(f"[CLEAN] Removing '{BUILD_DIR}'")
    shutil.rmtree(BUILD_DIR)


def generate_resources():
    cmd = "go generate resources/resources_gen.go"

    print(f'[GENERATE RESOURCES] Running "{cmd}"...')
    subprocess.run(cmd.split())


def prebuild():
    generate_resources()

    if not os.path.exists(BUILD_DIR):
        os.makedirs(BUILD_DIR)
        print(f"[PREBUILD] Created '{BUILD_DIR}'")

    command = "go mod tidy"
    print(f"[PREBUILD] Running '{command}'")
    subprocess.run(command.split())


def prerun():
    if not os.path.exists(LOGS_DIR):
        os.makedirs(LOGS_DIR)
        print(f"[PRERUN] Created '{LOGS_DIR}'")


def build_game(skip_prebuild=False):
    if not skip_prebuild:
        prebuild()

    print(
        f"[BUILD_GAME] Running 'go build -C {GAME_SRC} -o {BUILD_DIR}/game{ext} main.go'"
    )
    subprocess.run(f"go build -C {GAME_SRC} -o {BUILD_DIR}/game{ext} main.go".split())


def build_server(skip_prebuild=False):
    if not skip_prebuild:
        prebuild()

    print(
        f"[BUILD_SERVER] Running 'go build -C {SERVER_SRC} -o {BUILD_DIR}/server{ext} main.go'"
    )
    subprocess.run(
        f"go build -C {SERVER_SRC} -o {BUILD_DIR}/server{ext} main.go".split()
    )


def build():
    prebuild()
    build_game(skip_prebuild=True)
    build_server(skip_prebuild=True)


def run_game(skip_prerun=False, wait_for_exit=True, pipe_logs=True, skip_build=False):
    if not skip_prerun:
        prerun()
    if not skip_build:
        build_game()

    game_bin = os.path.normpath(f"{BUILD_DIR}/game{ext}")
    process = None

    print("Running game...")
    if pipe_logs:
        log_file = os.path.normpath(f"{LOGS_DIR}/game.log")
        with open(log_file, "+w") as log_file:
            process = subprocess.Popen([game_bin], stdout=log_file, stderr=log_file)
    else:
        process = subprocess.Popen([game_bin])

    if wait_for_exit:
        process.wait()


def run_server(
    skip_prerun=False, wait_for_exit=True, pipe_logs=True, skip_build=False
) -> subprocess.Popen[bytes] | None:
    if not skip_prerun:
        prerun()
    if not skip_build:
        build_server()

    server_bin = os.path.normpath(f"{BUILD_DIR}/server{ext}")
    process = None

    print("Running server...")
    if pipe_logs:
        log_file = os.path.normpath(f"{LOGS_DIR}/server.log")
        with open(log_file, "+w") as log_file:
            process = subprocess.Popen(server_bin, stdout=log_file, stderr=log_file)
    else:
        process = subprocess.Popen(
            server_bin,
        )

    if wait_for_exit:
        process.wait()
    else:
        return process

    return


def run():
    prerun()
    build()

    server_process = run_server(skip_prerun=True, wait_for_exit=False, skip_build=True)
    run_game(skip_prerun=True, skip_build=True)

    if server_process and os.name != "nt":
        server_process.send_signal(signal.SIGINT)
    elif os.name == "nt":
        server_process.kill()


def run_with_logs():
    prerun()
    build()

    server_process = run_server(
        skip_prerun=True, wait_for_exit=False, pipe_logs=False, skip_build=True
    )
    run_game(skip_prerun=True, pipe_logs=False, skip_build=True)

    if server_process and os.name != "nt":
        server_process.send_signal(signal.SIGINT)
    elif os.name == "nt":
        server_process.kill()

    print()


def format(root_path: str = "./src"):
    for root, dirs, files in os.walk(root_path):
        for file in filter(lambda x: x.endswith(".go"), files):
            subprocess.run(f"gofumpt -w -extra {root}/{file}".split())

        for dir in dirs:
            format(dir)


def lint():
    subprocess.run("golangci-lint run".split())


if len(sys.argv) > 1:
    action = sys.argv[1]

    match action:
        case "build":
            build()
        case "build_game":
            build_game()
        case "build_server":
            build_server()
        case "run":
            run()
        case "run_game":
            run_game()
        case "run_server":
            run_server()
        case "run_with_logs":
            run_with_logs()
        case "clean":
            clean()
        case "fmt" | "format":
            format()
        case "lint":
            lint()
        case _:
            print("Not a valid command")
            exit(1)
else:
    build()
