#!/usr/bin/env bash

BASE_FLAGS="-ldflags="

if [[ "$(uname)" == "Darwin" ]]; then
  echo "${BASE_FLAGS}-extldflags=-Wl,-ld_classic"
fi
