#!/usr/bin/env bash
# Run with the Go, clang and Node toolchain prepared by cloud/setup.sh.
set -euo pipefail
catalog=$(cd -- "$(dirname -- "$0")" && pwd)
exec python3 "$catalog/check.py" "$@"
