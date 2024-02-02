#!/usr/bin/env python3

import os
import shutil
import subprocess
import sys
import platform

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
GAME_SRC = os.path.normpath(f"{SRC_DIR}/game")
SERVER_SRC = os.path.normpath(f"{SRC_DIR}/server")

# if we're on wsl2, then we have to manually make sure that the GOOS environment variable is set to Linux
if platform.uname().release.endswith("microsoft-standard-WSL2"):
    os.environ["GOOS"] = "linux"

# server environment variables
SERVER_ENV = os.environ.copy()
SERVER_ENV["TCP_PORT"] = "8080"
SERVER_ENV["UDP_PORT"] = "8081"

# game environment variables
CLIENT_ENV = os.environ.copy()
CLIENT_ENV["SERVER_ADDRESS"] = "127.0.0.1"
CLIENT_ENV["SERVER_TCP_PORT"] = SERVER_ENV["TCP_PORT"]
CLIENT_ENV["SERVER_UDP_PORT"] = SERVER_ENV["UDP_PORT"]

ext = ".exe" if os.name == "nt" else ""


def clean():
    print(f"[CLEAN] Removing '{BUILD_DIR}'")
    shutil.rmtree(BUILD_DIR)


def prebuild():
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


def build_game():
    prebuild()

    print(
        f"[BUILD_GAME] Running 'go build -C {GAME_SRC} -o {BUILD_DIR}/game{ext} main.go'"
    )
    subprocess.run(f"go build -C {GAME_SRC} -o {BUILD_DIR}/game{ext} main.go".split())


def build_server():
    prebuild()
    
    print(
        f"[BUILD_SERVER] Running 'go build -C {SERVER_SRC} -o {BUILD_DIR}/server{ext} main.go'"
    )
    subprocess.run(
        f"go build -C {SERVER_SRC} -o {BUILD_DIR}/server{ext} main.go".split()
    )


def build():
    build_game()
    build_server()


def run_game(skip_prerun=False, wait=True, pipe_logs=True):
    if not skip_prerun:
        prerun()

    build_game()

    game_bin = os.path.normpath(f"{BUILD_DIR}/game{ext}")
    process = None

    print("Running game...")
    if pipe_logs:
        log_file = os.path.normpath(f"{LOGS_DIR}/game.log")
        with open(log_file, "+w") as log_file:
            process = subprocess.Popen(
                [game_bin], stdout=log_file, stderr=log_file, env=CLIENT_ENV
            )
    else:
        process = subprocess.Popen([game_bin], env=CLIENT_ENV)

    if wait:
        process.wait()


def run_server(skip_prerun=False, wait=True, pipe_logs=True):
    if not skip_prerun:
        prerun()

    build_server()

    server_bin = os.path.normpath(f"{BUILD_DIR}/server{ext}")
    process = None

    print("Running server...")
    if pipe_logs:
        log_file = os.path.normpath(f"{LOGS_DIR}/server.log")
        with open(log_file, "+w") as log_file:
            process = subprocess.Popen(
                [server_bin], stdout=log_file, stderr=log_file, env=SERVER_ENV
            )
    else:
        process = subprocess.Popen([server_bin], env=SERVER_ENV)

    if wait:
        process.wait()


def run():
    prerun()
    build()

    run_server(skip_prerun=True, wait=False)
    run_game(skip_prerun=True)


def run_with_logs():
    prerun()
    build()

    run_server(skip_prerun=True, wait=False, pipe_logs=False)
    run_game(skip_prerun=True, pipe_logs=False)
    print()


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
        case _:
            print("Not a valid command")
            exit(1)
else:
    build()
