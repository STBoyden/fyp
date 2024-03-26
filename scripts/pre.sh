#!/usr/bin/env bash

if [[ -f "/proc/sys/fs/binfmt_misc/WSLInterop" && $GOOS != "linux" ]]; then
  echo "You are using Linux under WSL, please make sure that you set GOOS to 'linux'. Current value: $GOOS"

  exit 127
fi
