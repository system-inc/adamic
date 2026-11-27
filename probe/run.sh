#!/usr/bin/env bash
set -eo pipefail
probe=$(cd -- "$(dirname -- "$0")" && pwd -P)
cd "$(git -C "$probe" rev-parse --show-toplevel)"
repo=$PWD
export TMPDIR="${LOOM_JOB_SCRATCH:-${TMPDIR:-/tmp}}"
source /opt/adamic-block/environment.sh &&
    export TMPDIR="${LOOM_JOB_SCRATCH:-$TMPDIR}" &&
    source cloud/sandboxed-job.sh &&
    sandboxed_repositories &&
    sandboxed_npm stage3/api &&
    sandboxed_typescript_source || exit 1
set -u

if [[ $# != 2 || ( $1 != base && $1 != head ) || ( $2 != entry && $2 != directory ) ]]; then
    echo "usage: $0 base|head entry|directory" >&2
    exit 2
fi
mode=$1
# One pass a job: both passes on both sides, about 30 minutes each, ran past the 2h ceiling (J-78s0qa8, Nov 25), and
# family 12 round 17's census (J-df4676b) took 1h59m48s.
selected=$2
base=45111cdcc9538e6b2d1f640b04e26c92b2baab13
if [[ $(git rev-parse HEAD) != "$base" ]]; then
    [[ $(git rev-parse --verify HEAD^) == "$base" ]] &&
        git diff --quiet "$base" HEAD -- . ':(exclude)probe' || {
        echo "Expected the pinned base at HEAD or at HEAD^ with only probe/ added" >&2
        exit 2
    }
fi

scratch=$(mktemp -d "${LOOM_JOB_SCRATCH:-$TMPDIR}/jsdoc-census-$mode.XXXXXX")
scratch=$(cd "$scratch" && pwd -P)
export TMPDIR="$scratch/tmp" GOTMPDIR="$scratch/tmp" GOCACHE="$scratch/go-cache"
export STAGE3_CACHE="$scratch/stage3-cache"
export ADAMIC_BUILD_CACHE_DIR="$scratch/products"
unset STAGE3_PRODUCT_STORE ADAMIC_BUILD_CACHE_URL GOCACHEPROG
source_tree="$scratch/source"
mkdir -p "$source_tree" "$TMPDIR"
printf 'mode=%s base=%s evidence=%s\n' "$mode" "$base" "$scratch"
git archive "$base" | tar -xf - -C "$source_tree"
# Go consumes the submodule sources already laid by the house job.
# The top-level revision check above holds the same gitlinks in both modes.
mkdir -p "$source_tree/cohere"
tar -C "$repo/cohere" --exclude=.git -cf - . | tar -xf - -C "$source_tree/cohere"
ln -s "$repo/stage3/api/node_modules" "$source_tree/stage3/api/node_modules"
cd "$source_tree"
if [[ $mode == head ]]; then
    git apply -p1 --exclude=stage3/upstream/LEDGER.md "$probe/change.patch"
fi

tree="$scratch/adapted"
if ! bash stage3/apply.sh "$tree" > "$scratch/apply.log" 2>&1; then
    tail -40 "$scratch/apply.log" >&2
    exit 1
fi

python3 stage3/census/latent/make_overlay.py "$PWD" "$scratch/latent-overlay"
# LATENT_TIMES logs completions. Add a start record before work begins
# in the scratch overlay so the unfinished unit is also observable.
python3 - "$PWD" "$scratch/latent-overlay" <<'PY'
import json
from pathlib import Path
import sys

source, overlay = map(Path, sys.argv[1:])
original = source / "stage3/census/latent/timing/timing.go"
replacement = overlay / "probe_timing.go"
text = original.read_text()
imports = '\t"encoding/json"\n'
start = '\twall, cpu := time.Now(), processCPUSeconds()\n'
assert text.count(imports) == text.count(start) == 1
text = text.replace(imports, imports + '\t"fmt"\n\t"os"\n')
text = text.replace(start, '\tfmt.Fprintf(os.Stderr, "unit-start\\t%s\\t%s\\n", file, where)\n' + start)
replacement.write_text(text)
manifest = overlay / "overlay.json"
data = json.loads(manifest.read_text())
assert str(original) not in data["Replace"]
data["Replace"][str(original)] = str(replacement)
manifest.write_text(json.dumps(data) + "\n")
PY
gofmt -w "$scratch/latent-overlay/"*.go
python3 stage3/meter/entry_overlay.py "$scratch/latent-overlay" "$scratch/entry-overlay"
gofmt -w "$scratch/entry-overlay/"*.go
go build -buildvcs=false -overlay="$scratch/entry-overlay/overlay.json" -o "$scratch/entry-census" ./stage3/census/latent/tool
go build -buildvcs=false -overlay="$scratch/latent-overlay/overlay.json" -o "$scratch/directory-census" ./stage3/census/latent/tool

pass() {
    local name=$1 binary=$2 target=$3 began=$SECONDS rc=0 summary_rc=0
    local prefix="$scratch/$name"
    env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 CENSUS_MODULE_FORMAT=CommonJSBundle \
        LATENT_TIMES="$prefix.times.jsonl" timeout --signal=KILL 50m "$binary" "$target" "$prefix.jsonl" \
        > "$prefix.log" 2>&1 || rc=$?
    printf '%s wall_seconds=%s exit=%s\n' "$name" "$((SECONDS - began))" "$rc"
    if (( rc != 0 )); then
        printf '%s last units started (up to 10)\n' "$name"
        awk -F '\t' '$1=="unit-start" {print $3}' "$prefix.log" | tail -10
    fi
    python3 - "$tree" "$prefix" "$rc" <<'PY' || summary_rc=$?
import hashlib
import json
from pathlib import Path
import re
import sys

tree, prefix = sys.argv[1:3]
rc = int(sys.argv[3])
timings = Path(prefix + ".times.jsonl")
completed = 0
if timings.exists():
    for line in timings.read_text().splitlines():
        try:
            json.loads(line)
        except json.JSONDecodeError:
            break
        completed += 1
print(f"completed_units={completed}")
stream = Path(prefix + ".jsonl")
rows = []
if stream.exists():
    for line in stream.read_text().splitlines():
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            break
complete = (rc == 0 and len(rows) >= 3 and rows[0].get("status") == "measurement"
            and not rows[0].get("checker_rejected") and not rows[0].get("diagnostics")
            and rows[0].get("latent_mode") == "first-error"
            and rows[0].get("module_format") == "CommonJSBundle"
            and rows[-1].get("status") == "relation_report")
if complete:
    files = {row["file"] for row in rows if "file" in row}
    expected = set(rows[0].get("sources", []))
    if not expected:
        expected = {str(p) for p in (Path(tree) / "src/compiler").rglob("*") if p.is_file() and p.suffix in {".ts", ".a"}}
    complete = files == expected and len(files) == sum("file" in row for row in rows)
position = re.compile(r"(\.[cm]?ts|\.a)(?::\d+(?::\d+)?|\(\d+,\s*\d+\))")
def normalize(text):
    return position.sub(r"\1", text.replace(tree + "/", ""))
findings = [finding for row in rows for finding in row.get("findings") or []]
# The meter counts distinct (kind, where, reason, text), with scoped tail separate.
for label, tail in (("sites", False), ("tail-sites", True)):
    sites = {(f["kind"], f["where"].replace(tree + "/", ""), normalize(f["reason"]), normalize(f["text"]))
             for f in findings if (f.get("phase") == "tail") == tail}
    path = Path(prefix + "." + label + (".jsonl" if complete else ".partial.jsonl"))
    data = "".join(json.dumps(site, ensure_ascii=False) + "\n" for site in sorted(sites)).encode()
    path.write_bytes(data)
    state = "complete" if complete else "incomplete; overall count unavailable"
    print(f"{label} observed_count={len(sites)} sha256={hashlib.sha256(data).hexdigest()} {state} path={path}")
sys.exit(0 if complete else 1)
PY
    (( rc == 0 && summary_rc == 0 ))
}
result=0
if [[ $selected == entry ]]; then
    pass entry "$scratch/entry-census" "$tree/src/tsc/tsc.ts" || result=1
else
    pass directory "$scratch/directory-census" "$tree/src/compiler" || result=1
fi
exit "$result"
