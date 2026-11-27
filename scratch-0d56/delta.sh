#!/bin/bash
# #0d56xpd: the admission delta of the head against its parent, on a house box whose lay kit lays cohere as a
# repository. Scratch only: never lands. Verdict lines begin "d56:".
set -u
source ~/adamic-tools/env.sh > /dev/null 2>&1 && echo "d56: PASS env.sh sourced" || echo "d56: FAIL env.sh"
export GOFLAGS=-buildvcs=false
mkdir -p "${TMPDIR:=/tmp/d56}" && echo "d56: PASS scratch $TMPDIR" || { echo "d56: FAIL scratch $TMPDIR"; exit 1; }
git cat-file -e "65cbf511b8981752c861dadf0e73fd9993bf817e^{commit}" && echo "d56: PASS base present" || echo "d56: FAIL base not in this clone"
head=2a2360270bc8f9b328a8c8aa6f8457c5f481b87f
base=65cbf511b8981752c861dadf0e73fd9993bf817e
echo "d56: head $head base $base"
python3 cloud/admission-corpus/manifest.py --sha "$head" > "$TMPDIR/manifest.json" 2> "$TMPDIR/manifest.err" || { tail -5 "$TMPDIR/manifest.err"; echo "d56: verdict manifest FAIL"; }
go run ./cmd/adamic-admission-delta --base "$base" --head "$head" --manifest "$TMPDIR/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$TMPDIR/delta.json" 2> "$TMPDIR/delta.err"
rc=$?
tail -8 "$TMPDIR/delta.err" | cut -c1-300
python3 -I - "$TMPDIR/delta.json" <<'PYTHON'
import json, sys
try:
    data = json.load(open(sys.argv[1]))
except Exception as error:
    print(f"d56: delta json unreadable: {error}")
    sys.exit(0)
def walk(node, path=""):
    if isinstance(node, dict):
        for key, value in node.items():
            walk(value, f"{path}.{key}")
    elif isinstance(node, list):
        if any(word in path.lower() for word in ("admitted", "newly", "accepted", "refus", "fail", "diverg")):
            print(f"d56: {path}: {len(node)}")
            for item in node[:40]:
                print(f"d56:   {json.dumps(item)[:240]}")
        for item in node:
            walk(item, path)
walk(data)
def diffs(node):
    if isinstance(node, dict):
        if any(value == "diff" for value in node.values()):
            print(f"d56: diff {json.dumps(node)[:900]}")
        for value in node.values():
            diffs(value)
    elif isinstance(node, list):
        for item in node:
            diffs(item)
diffs(data)
print(f"d56: keys {sorted(data.keys()) if isinstance(data, dict) else type(data)}")
summary = {k: v for k, v in data.items() if not isinstance(v, (list, dict))} if isinstance(data, dict) else {}
print(f"d56: summary {json.dumps(summary)[:600]}")
PYTHON
echo "d56: verdict admission delta exit=$rc"
