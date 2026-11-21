#!/bin/bash
# #rnbvya6: the admission delta of the fix against main, on a house box whose lay kit lays cohere as a repository.
# Scratch only: never lands. Verdict lines begin "rnbd:".
set -u
source ~/adamic-tools/env.sh > /dev/null 2>&1
export GOFLAGS=-buildvcs=false
mkdir -p "${TMPDIR:=/tmp/rnbd}"
head=51c69dc05e5de8e530b4cac770657fa602d17d08
base=f06a6d883c2fbbf8e63d0967158d65b1bea21547
echo "rnbd: head $head base $base"
python3 cloud/admission-corpus/manifest.py --sha "$head" > "$TMPDIR/manifest.json" 2> "$TMPDIR/manifest.err" || { tail -5 "$TMPDIR/manifest.err"; echo "rnbd: verdict manifest FAIL"; }
go run ./cmd/adamic-admission-delta --base "$base" --head "$head" --manifest "$TMPDIR/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$TMPDIR/delta.json" 2> "$TMPDIR/delta.err"
rc=$?
tail -8 "$TMPDIR/delta.err" | cut -c1-300
python3 -I - "$TMPDIR/delta.json" <<'PYTHON'
import json, sys
try:
    data = json.load(open(sys.argv[1]))
except Exception as error:
    print(f"rnbd: delta json unreadable: {error}")
    sys.exit(0)
def walk(node, path=""):
    if isinstance(node, dict):
        for key, value in node.items():
            walk(value, f"{path}.{key}")
    elif isinstance(node, list):
        if any(word in path.lower() for word in ("admitted", "newly", "accepted", "refus", "fail", "diverg")):
            print(f"rnbd: {path}: {len(node)}")
            for item in node[:40]:
                print(f"rnbd:   {json.dumps(item)[:240]}")
        for item in node:
            walk(item, path)
walk(data)
summary = {k: v for k, v in data.items() if not isinstance(v, (list, dict))} if isinstance(data, dict) else {}
print(f"rnbd: summary {json.dumps(summary)[:600]}")
PYTHON
echo "rnbd: verdict admission delta exit=$rc"
