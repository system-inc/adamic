#!/usr/bin/env bash
# Run with the Go, clang and Node toolchain prepared by cloud/setup.sh.
set -euo pipefail
if [ "$#" -ne 1 ]; then
    echo 'usage: verify/catalog/check.sh <commit>' >&2
    exit 2
fi
catalog=$(cd -- "$(dirname -- "$0")" && pwd)
exec python3 "$catalog/check.py" "$1"
