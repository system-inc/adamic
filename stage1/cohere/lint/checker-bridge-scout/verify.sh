#!/usr/bin/env bash
# Caller supplies a fresh scratch directory and the repository's Go/clang/Node tool PATH.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)
package="$root/stage1/cohere/lint/checker-bridge-scout"
task_dir=${1:?usage: verify.sh /absolute/new/scratch/directory}
mkdir "$task_dir"
cd "$root"
python3 "$package/fetch_public.py" "$task_dir/public" > "$task_dir/fetch.log" 2>&1
python3 "$package/sample_public.py" "$task_dir/public" > "$task_dir/sample.log" 2>&1
corpus="$task_dir/public/microsoft__TypeScript-050880ce"
mkdir -p "$corpus/src/compiler"
cp "$task_dir/public/samples/microsoft__TypeScript-050880ce/src/compiler/core.ts" "$corpus/src/compiler/core.ts"
go build -o "$task_dir/adamic" ./cmd/adamic > "$task_dir/compiler-build.log" 2>&1
go build -buildmode=c-archive -o "$task_dir/tsgo.a" ./bridge/tsgo/archive > "$task_dir/archive-build.log" 2>&1
CC=clang CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all' \
 go build -buildmode=c-archive -o "$task_dir/tsgo-asan.a" ./bridge/tsgo/archive > "$task_dir/archive-asan-build.log" 2>&1
for name in driver pilot; do
 "$task_dir/adamic" build "$package/$name.ts" -o "$task_dir/$name" --tsgo "$task_dir/tsgo.a" > "$task_dir/$name-build.log" 2>&1
 "$task_dir/adamic" build "$package/$name.ts" -o "$task_dir/$name-asan" --tsgo "$task_dir/tsgo-asan.a" --sanitize > "$task_dir/$name-asan-build.log" 2>&1
done
python3 - "$root" "$task_dir" <<'PY'
import json, pathlib, sys
root, task = map(pathlib.Path, sys.argv[1:])
virtual = root/'cohere/adamic_scout_oracle.go'
source = root/'stage1/cohere/lint/checker-bridge-scout/testdata/oracle.go'
(task/'overlay.json').write_text(json.dumps({'Replace': {str(virtual): str(source)}}))
PY
go -C cohere build -overlay="$task_dir/overlay.json" -o "$task_dir/oracle" "$root/cohere/adamic_scout_oracle.go" > "$task_dir/oracle-build.log" 2>&1
ADAMIC_SCOUT_NATIVE="$task_dir/driver-asan" ADAMIC_SCOUT_PILOT="$task_dir/pilot-asan" \
 ADAMIC_TYPESCRIPT_SOURCE="$corpus" ADAMIC_SCOUT_PUBLIC_MANIFEST="$task_dir/public/samples/manifest" \
 ASAN_OPTIONS=detect_leaks=1 GOMAXPROCS=4 \
 go test -v -count=1 -timeout 30m ./stage1/cohere/lint/checker-bridge-scout > "$task_dir/tests.log" 2>&1
go vet ./stage1/cohere/lint/checker-bridge-scout > "$task_dir/vet.log" 2>&1
gofmt -l "$package"/*.go "$package/testdata/oracle.go" > "$task_dir/gofmt.log"
test ! -s "$task_dir/gofmt.log"
go build -o "$task_dir/measure" ./stage1/cohere/lint/checker-bridge-scout
config="$task_dir/public/samples/tsconfig.json"
manifest="$task_dir/public/samples/manifest"
for repetitions in 1 100; do
 GOMAXPROCS=1 "$task_dir/measure" -config "$config" -manifest "$manifest" -native "$task_dir/driver" \
  -pilot "$task_dir/pilot" -oracle "$task_dir/oracle" -out "$task_dir/perfile-$repetitions" -repetitions "$repetitions" -rounds 3 > "$task_dir/perfile-$repetitions.log" 2>&1
 GOMAXPROCS=1 python3 "$package/measure_whole.py" "$task_dir/pilot" "$task_dir/oracle" "$config" "$manifest" \
  "$task_dir/whole-$repetitions" --repetitions "$repetitions" --rounds 3 > "$task_dir/whole-$repetitions.log" 2>&1
done
GOMAXPROCS=1 "$task_dir/measure" -config "$config" -manifest "$manifest" -native "$task_dir/driver" \
 -pilot "$task_dir/pilot" -oracle "$task_dir/oracle" -out "$task_dir/profile" -repetitions 100 -rounds 1 -profile > "$task_dir/profile.log" 2>&1
GOMAXPROCS=1 python3 "$package/measure_whole.py" "$task_dir/pilot" "$task_dir/oracle" "$config" "$manifest" \
 "$task_dir/whole-profile" --repetitions 100 --rounds 1 --profile > "$task_dir/whole-profile.log" 2>&1
# Full physical typed-source census and owner/cache continuation; no repo code runs.
python3 "$package/materialize_public.py" "$task_dir/public/pins.json" > "$task_dir/materialize.log" 2>&1
GOMAXPROCS=4 "$task_dir/measure" census "$task_dir/public/pins.json" "$task_dir/census.jsonl"
python3 "$package/summarize_census.py" "$task_dir/public/pins.json" "$task_dir/census.jsonl" "$task_dir/census-summary.json"
python3 "$package/measure_continuation.py" "$task_dir/measure" "$task_dir/driver" \
 "$config" "$manifest" "$task_dir/continuation"

GOMAXPROCS=4 "$task_dir/measure" batches "$config" "$manifest" "$task_dir/batches.json"
