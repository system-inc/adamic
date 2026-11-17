#!/bin/bash
# c5 (#snypn8c): what lowering admits moved, so oracle whole, the latent census, native with WASI on, and the
# admission delta against the head's parent. Scratch only: never lands. Verdict lines begin "c5w: verdict".
set -u
# The sealed helpers are shell functions, which a child bash doesn't inherit, so they're sourced here.
source /opt/adamic-block/environment.sh
source cloud/sealed-job.sh
v() { echo "c5w: verdict $*"; }
head=$(git rev-parse HEAD~1)
base=$(git rev-parse HEAD~2)
echo "c5w: head $head base $base"
sealed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 && v "npm PASS" || { tail -5 "$TMPDIR/npm.log"; v "npm FAIL"; }
sealed_products ./internal/oracle ./internal/native > "$TMPDIR/products.log" 2>&1 && v "products PASS" || { tail -10 "$TMPDIR/products.log"; v "products FAIL"; }
go test -count=1 -timeout 45m ./internal/oracle/ > "$TMPDIR/oracle.log" 2>&1
rc=$?
grep -E "^(--- FAIL|FAIL|ok |panic:)" "$TMPDIR/oracle.log" | cut -c1-240 | head -30
v "internal/oracle whole exit=$rc"
ADAMIC_TEST_WASI=1 go test -count=1 -timeout 45m ./internal/native/ > "$TMPDIR/native.log" 2>&1
rc=$?
grep -E "^(--- FAIL|FAIL|ok |panic:)" "$TMPDIR/native.log" | cut -c1-240 | head -30
v "internal/native ADAMIC_TEST_WASI=1 exit=$rc"
python3 cloud/admission-corpus/manifest.py --sha "$head" > "$TMPDIR/manifest.json" 2> "$TMPDIR/manifest.err" || { tail -5 "$TMPDIR/manifest.err"; v "delta manifest FAIL"; }
go run ./cmd/adamic-admission-delta --base "$base" --head "$head" --manifest "$TMPDIR/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$TMPDIR/delta.json" 2> "$TMPDIR/delta.err"
rc=$?
tail -8 "$TMPDIR/delta.err" | cut -c1-300
python3 -I - "$TMPDIR/delta.json" <<'PYTHON'
import json, sys
try:
    data = json.load(open(sys.argv[1]))
except Exception as error:
    print(f"c5w: delta json unreadable: {error}")
    sys.exit(0)
def walk(node, path=""):
    if isinstance(node, dict):
        for key, value in node.items():
            walk(value, f"{path}.{key}")
    elif isinstance(node, list):
        if any(word in path.lower() for word in ("admitted", "newly", "accepted", "refus", "fail", "diverg")):
            print(f"c5w: delta {path}: {len(node)}")
            for item in node[:40]:
                print(f"c5w:   {json.dumps(item)[:220]}")
        for item in node:
            walk(item, path)
walk(data)
summary = {k: v for k, v in data.items() if not isinstance(v, (list, dict))} if isinstance(data, dict) else {}
print(f"c5w: delta summary {json.dumps(summary)[:600]}")
PYTHON
v "admission delta exit=$rc"
