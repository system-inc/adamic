#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
# Source cloud/setup.sh's printed environment before invoking this script.
: "${WASI_SYSROOT:?source the WASI setup environment first}"
python3 bench/wasm/run.py "$@"
