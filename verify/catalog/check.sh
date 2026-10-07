#!/usr/bin/env bash
# Run with the Go, clang and Node toolchain prepared by cloud/setup.sh.
set -euo pipefail
catalog=$(cd -- "$(timeout --verbose --kill-after=1s 30 dirname -- "$0")" && pwd)
source "$catalog/../../internal/boundedrun/shell.sh"
# Entries are parallel, but a complete catalog can take substantially longer than
# one Go test. The per-child helper retains Go's 70m ceiling and 10m for Git.
bounded 14400 python3 "$catalog/../../internal/boundedrun/python.py" "$catalog/check.py" "$@"
