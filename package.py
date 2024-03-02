#!/usr/bin/env python3

import os
import argparse
import platform
import shutil
import subprocess

if __name__ != "__main__":
    print("This script needs to be ran directly")
    exit(1)


SRC_DIR = "src/cmd/client"


def get_default_platform() -> str:
    if os.name == "nt":
        return "windows"
    else:
        if "Linux" in platform.platform():
            return "linux"
        else:
            return "macos"


default_platform = get_default_platform()

parser = argparse.ArgumentParser(
    prog="package.py",
    description="Builds and bundles the game for a given platform.",
    epilog="Copyright (c) 2024 Samuel Boyden",
)

parser.add_argument(
    "-p",
    "--platform",
    type=str,
    choices=["windows", "macos", "linux"],
    default=default_platform,
    dest="target_platform",
    help=f"the target platform to compile and bundle for (default: {default_platform})",
)

args = parser.parse_args()

if os.path.exists("dist"):
    shutil.rmtree("dist")

os.makedirs("dist")
shutil.copytree("resources", "dist/resources", dirs_exist_ok=True)
shutil.copyfile(".env.example", "dist/.env")

extension = ".exe" if args.target_platform == "windows" else ""
build_command = f"go build -o dist/game{extension} {SRC_DIR}/main.go"

env = os.environ.copy()
env["GOOS"] = args.target_platform if args.target_platform != "macos" else "darwin"

print(f'Running "{build_command}" with GOOS env as "{env["GOOS"]}"...')

subprocess.run(build_command.split(), env=env)

archive_name = f"fyp-build-{args.target_platform}.zip"

if os.path.isfile(archive_name):
    os.remove(archive_name)

compress_command = (
    f"powershell Compress-Archive . ../{archive_name}"
    if os.name == "nt"
    else f"zip -r ../{archive_name} ."
)

print(f'Running "{compress_command}"...')

subprocess.run(compress_command.split(), cwd="dist")
