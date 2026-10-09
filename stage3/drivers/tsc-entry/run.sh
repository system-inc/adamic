#!/usr/bin/env bash
# Run with main's compiler checkout as cwd and its toolchain env.sh sourced.
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
out=${1:?usage: run.sh NEW_OUTPUT_DIRECTORY}
mkdir "$out"
out=$(cd "$out" && pwd)
node --version > "$out/node-version.log"
[ "$(node --version)" = v24.19.0 ]
git rev-parse HEAD > "$out/compiler-commit.log"
bash "$here/../../apply.sh" "$out/adapted" > "$out/apply.log" 2>&1
api=${STAGE3_CACHE:-$HOME/.cache/adamic-stage3}/api/node_modules/typescript/lib/typescript.js
SCANNER_TYPESCRIPT="$api" node "$here/graph.cjs" "$out/adapted" "$out" > "$out/graph.log" 2>&1
go build -o "$out/adamic" ./cmd/adamic > "$out/compiler-build.log" 2>&1
for split in 0 1; do
    status=0
    ADAMIC_NATIVE_SPLIT=$split ADAMIC_NATIVE_JOBS=$(nproc) "$out/adamic" build "$out/adapted/src/tsc/tsc.ts" -o "$out/tsc-$split" > "$out/build-$split.stdout" 2> "$out/build-$split.stderr" || status=$?
    echo "$status" > "$out/build-$split.exit"
done
