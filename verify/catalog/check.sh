#!/usr/bin/env bash
# Run with the Go, clang and Node toolchain prepared by cloud/setup.sh.
set -euo pipefail
scriptDirectory=${BASH_SOURCE[0]%/*}
[ "$scriptDirectory" != "${BASH_SOURCE[0]}" ] || scriptDirectory=.
catalog=$(cd -- "$scriptDirectory" && pwd)
source "$catalog/../../internal/boundedrun/shell.sh"
# Entries are parallel, but a complete catalog can take substantially longer than
# one Go test. The per-child helper retains Go's 70m ceiling and 10m for Git.
bounded 14400 python3 "$catalog/../../internal/boundedrun/python.py" "$catalog/check.py" "$@"
