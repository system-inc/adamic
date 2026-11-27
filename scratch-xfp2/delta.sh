#!/bin/bash
# #xfpsqtn (1): the admission delta of the head against its parent, on a house box whose lay kit lays cohere as a
# repository. Scratch only: never lands. Verdict lines begin "xfpd:".
set -u
mkdir -p "${TMPDIR:=/tmp/xfpd}" && echo "xfpd: PASS mkdir $TMPDIR" || { echo "xfpd: FAIL mkdir $TMPDIR"; exit 1; }
source cloud/sandboxed-job.sh > "$TMPDIR/sandboxed-job.log" 2>&1 && echo "xfpd: PASS sandboxed-job" || { echo "xfpd: FAIL sandboxed-job"; tail -5 "$TMPDIR/sandboxed-job.log"; }
source ~/adamic-tools/env.sh > /dev/null 2>&1 && echo "xfpd: PASS adamic-tools env" || echo "xfpd: FAIL adamic-tools env"
export GOFLAGS=-buildvcs=false
head=f98db72abaf31457a5520be3954e53a5104e2c98
base=65cbf511b8981752c861dadf0e73fd9993bf817e
echo "xfpd: head $head base $base"
python3 cloud/admission-corpus/manifest.py --sha "$head" > "$TMPDIR/manifest.json" 2> "$TMPDIR/manifest.err" || { tail -5 "$TMPDIR/manifest.err"; echo "xfpd: verdict manifest FAIL"; }
go run ./cmd/adamic-admission-delta --base "$base" --head "$head" --manifest "$TMPDIR/manifest.json" \
	--manifest-generator cloud/admission-corpus/manifest.py --json > "$TMPDIR/delta.json" 2> "$TMPDIR/delta.err"
rc=$?
tail -8 "$TMPDIR/delta.err" | cut -c1-300
python3 -I - "$TMPDIR/delta.json" <<'PYTHON'
import json, sys
try:
    data = json.load(open(sys.argv[1]))
except Exception as error:
    print(f"xfpd: delta json unreadable: {error}")
    sys.exit(0)
def walk(node, path=""):
    if isinstance(node, dict):
        for key, value in node.items():
            walk(value, f"{path}.{key}")
    elif isinstance(node, list):
        if any(word in path.lower() for word in ("admitted", "newly", "accepted", "refus", "fail", "diverg")):
            print(f"xfpd: {path}: {len(node)}")
            for item in node[:40]:
                print(f"xfpd:   {json.dumps(item)[:240]}")
        for item in node:
            walk(item, path)
walk(data)
def diffs(node):
    if isinstance(node, dict):
        if any(value == "diff" for value in node.values()):
            print(f"xfpd: diff {json.dumps(node)[:900]}")
        for value in node.values():
            diffs(value)
    elif isinstance(node, list):
        for item in node:
            diffs(item)
diffs(data)
print(f"xfpd: keys {sorted(data.keys()) if isinstance(data, dict) else type(data)}")
summary = {k: v for k, v in data.items() if not isinstance(v, (list, dict))} if isinstance(data, dict) else {}
print(f"xfpd: summary {json.dumps(summary)[:600]}")
PYTHON
echo "xfpd: verdict admission delta exit=$rc"
