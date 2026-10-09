#!/usr/bin/env bash
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
if [[ $# != 4 ]]; then
    echo 'usage: run.sh CANDIDATE_WORKTREE SCOUT_DIRECTORY ADAPTED_TREE NEW_OUTPUT' >&2
    exit 2
fi
candidate=$(realpath "$1")
scout=$(realpath "$2")
tree=$(realpath "$3")
out=$(realpath -m "$4")
if [[ -e "$out" ]]; then echo "output must be new: $out" >&2; exit 2; fi
: "${STEP31_TYPESCRIPT:?stock typescript 6.0.3 API pathname required}"
mkdir -p "$out"
(
    cd "$candidate"
    go build -buildvcs=false -o "$out/adamic" ./cmd/adamic
) > "$out/compiler.log" 2>&1
node "$here/record.mjs" "$tree" "$out/resolver.json" "$scout" > "$out/record.log" 2>&1
python3 - "$out/resolver.json" "$out/request.json" <<'PY'
import json,sys
from pathlib import Path
record=json.loads(Path(sys.argv[1]).read_text())
Path(sys.argv[2]).write_text(json.dumps(record['request'],separators=(',',':'))+'\n')
PY
SLICE_TYPESCRIPT=$STEP31_TYPESCRIPT bash "$candidate/stage3/slice/run.sh" "$tree" "$out/slice" \
    src/compiler/emitter.ts:emitFiles src/compiler/parser.ts:createSourceFile \
    src/compiler/binder.ts:bindSourceFile src/compiler/transformer.ts:getTransformers \
    src/compiler/core.ts:createMultiMap --no-adapt > "$out/slice.log" 2>&1
node "$candidate/stage3/slice/verify.cjs" "$out/slice" > "$out/verify.log" 2>&1
node "$here/make-driver.cjs" "$out/resolver.json" "$tree" "$out/slice" "$out/replay.a" > "$out/driver.log" 2>&1
node "$scout/emitter-dump.cjs" "$out/request.json" "$out/stock" > "$out/stock.log" 2>&1
# This establishes component evidence on Node. It never occupies a native slot.
python3 "$scout/component-compare.py" "$out/stock" "$out/node-comparison" -- \
    node "$here/node.mjs" "$out/replay.a" > "$out/node-compare.log" 2>&1
blocked=0
for split in 0 1; do
    set +e
    (cd "$candidate"; ADAMIC_NATIVE_SPLIT=$split "$out/adamic" build "$out/replay.a" -o "$out/emitter-$split") \
        > "$out/native-$split.stdout" 2> "$out/native-$split.stderr"
    code=$?
    set -e
    printf '%s\n' "$code" > "$out/native-$split.exit"
    if [[ $code == 0 ]]; then
        python3 "$scout/component-compare.py" "$out/stock" "$out/native-$split-comparison" -- \
            "$out/emitter-$split" > "$out/native-$split-compare.log" 2>&1
    else
        blocked=$((blocked+1))
    fi
done
if [[ $blocked == 2 ]]; then
    python3 "$here/walk.py" "$out/adamic" "$candidate" "$out/slice" "$out/replay.a" "$out/walk" > "$out/walk.log" 2>&1
    node "$here/map-stops.cjs" "$out/slice" "$out/walk" "$tree" "$out/stops.json" > "$out/map.log" 2>&1
fi
python3 "$here/test-witnesses.py" "$out/adamic" "$candidate" "$out/witnesses" > "$out/witnesses.log" 2>&1
python3 "$here/test-replay.py" "$scout" "$out/resolver.json" "$tree" "$out/slice" "$out/stock" "$out/replay-mutants" > "$out/replay-mutants.log" 2>&1
printf 'PASS: component Node comparison, both native attempts, bounded stops, witnesses and mutants\n'
