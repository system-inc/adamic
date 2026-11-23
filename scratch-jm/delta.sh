#!/bin/bash
# #jm2ba55: the admission delta of the head against its parent, on a house box whose lay kit lays cohere as a
# repository. Scratch only: never lands. Verdict lines begin "jmd:".
set -u
source ~/adamic-tools/env.sh > /dev/null 2>&1
export GOFLAGS=-buildvcs=false
mkdir -p "${TMPDIR:=/tmp/jmd}"
head=ad5d99748f6fffdff3d650c4ab93941947c93623
base=fb672853cf6aeddffe2e2de7d4a3ae95b0df3215
echo "jmd: head $head base $base"
python3 cloud/admission-corpus/manifest.py --sha "$head" > "$TMPDIR/manifest.json" 2> "$TMPDIR/manifest.err" || { tail -5 "$TMPDIR/manifest.err"; echo "jmd: verdict manifest FAIL"; }
go run ./cmd/adamic-admission-delta --base "$base" --head "$head" --manifest "$TMPDIR/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$TMPDIR/delta.json" 2> "$TMPDIR/delta.err"
rc=$?
tail -8 "$TMPDIR/delta.err" | cut -c1-300
python3 -I - "$TMPDIR/delta.json" <<'PYTHON'
import json, sys
try:
    data = json.load(open(sys.argv[1]))
except Exception as error:
    print(f"jmd: delta json unreadable: {error}")
    sys.exit(0)
def walk(node, path=""):
    if isinstance(node, dict):
        for key, value in node.items():
            walk(value, f"{path}.{key}")
    elif isinstance(node, list):
        if any(word in path.lower() for word in ("admitted", "newly", "accepted", "refus", "fail", "diverg")):
            print(f"jmd: {path}: {len(node)}")
            for item in node[:40]:
                print(f"jmd:   {json.dumps(item)[:240]}")
        for item in node:
            walk(item, path)
walk(data)
def diffs(node):
    if isinstance(node, dict):
        if any(value == "diff" for value in node.values()):
            print(f"jmd: diff {json.dumps(node)[:900]}")
        for value in node.values():
            diffs(value)
    elif isinstance(node, list):
        for item in node:
            diffs(item)
diffs(data)
print(f"jmd: keys {sorted(data.keys()) if isinstance(data, dict) else type(data)}")
summary = {k: v for k, v in data.items() if not isinstance(v, (list, dict))} if isinstance(data, dict) else {}
print(f"jmd: summary {json.dumps(summary)[:600]}")
PYTHON
echo "jmd: verdict admission delta exit=$rc"
