#!/bin/bash
# #0q0cv0m: the admission delta of the head against main 25f3ed7f, on a house box whose lay kit lays cohere as a
# repository. Scratch only: never lands. Verdict lines begin "0q0d:".
set -u
source ~/adamic-tools/env.sh > /dev/null 2>&1
source cloud/sandboxed-job.sh && echo "0q0d: PASS sandboxed-job" || echo "0q0d: FAIL sandboxed-job"
export GOFLAGS=-buildvcs=false
mkdir -p "${TMPDIR:=/tmp/0q0d}/scratch-0q0d" && echo "0q0d: PASS scratch" || { echo "0q0d: FAIL scratch"; exit 1; }
scratch="$TMPDIR/scratch-0q0d"
head=217b327989028054c5a5c38c216394963a6dc7d4
base=25f3ed7f485abd8a6e7db292b70b7af4523fcdd2
echo "0q0d: head $head base $base"
python3 cloud/admission-corpus/manifest.py --sha "$head" > "$scratch/manifest.json" 2> "$scratch/manifest.err" && echo "0q0d: PASS manifest" || { tail -5 "$scratch/manifest.err"; echo "0q0d: FAIL manifest"; }
go run ./cmd/adamic-admission-delta --base "$base" --head "$head" --manifest "$scratch/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$scratch/delta.json" 2> "$scratch/delta.err"
rc=$?
tail -8 "$scratch/delta.err" | cut -c1-300
python3 -I - "$scratch/delta.json" <<'PYTHON'
import json, sys
try:
    data = json.load(open(sys.argv[1]))
except Exception as error:
    print(f"0q0d: delta json unreadable: {error}")
    sys.exit(0)
def walk(node, path=""):
    if isinstance(node, dict):
        for key, value in node.items():
            walk(value, f"{path}.{key}")
    elif isinstance(node, list):
        if any(word in path.lower() for word in ("admitted", "newly", "accepted", "refus", "fail", "diverg", "moved", "changed")):
            print(f"0q0d: {path}: {len(node)}")
            for item in node[:40]:
                print(f"0q0d:   {json.dumps(item)[:240]}")
        for item in node:
            walk(item, path)
walk(data)
def diffs(node):
    if isinstance(node, dict):
        if any(value == "diff" for value in node.values()):
            print(f"0q0d: diff {json.dumps(node)[:900]}")
        for value in node.values():
            diffs(value)
    elif isinstance(node, list):
        for item in node:
            diffs(item)
diffs(data)
print(f"0q0d: keys {sorted(data.keys()) if isinstance(data, dict) else type(data)}")
summary = {k: v for k, v in data.items() if not isinstance(v, (list, dict))} if isinstance(data, dict) else {}
print(f"0q0d: summary {json.dumps(summary)[:600]}")
PYTHON
echo "0q0d: verdict admission delta exit=$rc"
