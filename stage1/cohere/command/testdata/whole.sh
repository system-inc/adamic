#!/usr/bin/env bash
# Reproduce only in scratch: no shared generated/source file is written.
set -euo pipefail
repository=$(git rev-parse --show-toplevel)
scratch=$(mktemp -d "${TMPDIR:-/tmp}/scout42-whole.XXXXXX")
mkdir -p "$scratch/stage1" "$scratch/tools"
cp -a "$repository/stage1/cohere" "$scratch/stage1/"
cp -a "$repository/stage1/typescript" "$scratch/stage1/"
cp "$repository/stage1/cohere/command/testdata/whole.a.txt" "$scratch/stage1/cohere/command/probe.a"
cd "$repository"
go run ./cmd/lint-registry -root "$scratch/stage1/cohere/lint" > "$scratch/registry.log"
go build -buildmode=c-archive -o "$scratch/checker.a" ./bridge/tsgo/archive
cp stage1/cohere/command/testdata/measure_clang.py "$scratch/tools/clang"
chmod +x "$scratch/tools/clang"
PATH="$scratch/tools:$PATH" go run stage1/cohere/command/testdata/measure_whole.go "$scratch/stage1/cohere/command/probe.a" "$scratch/checker.a" "$scratch/cohere"
node --disable-warning=ExperimentalWarning oracle/node.mjs "$scratch/stage1/cohere/command/probe.a" > "$scratch/node.txt"
"$scratch/cohere" > "$scratch/native.txt"
cmp "$scratch/node.txt" "$scratch/native.txt"
go -C cohere test -count=1 -run '^TestNoDebuggerFires$' ./internal/lint/rules/core
printf '%s\n' "$scratch"
