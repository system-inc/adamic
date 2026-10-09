#!/usr/bin/env bash
# Run from repository root on the main SHA recorded in evidence/provenance.json.
# Check the active census log at least once a minute. Every command is bounded.
set -euo pipefail
scratch=${1:?usage: bash reproduce.sh NEW_SCRATCH_DIRECTORY}
mkdir "$scratch"
scratch=$(realpath "$scratch")
repo=$PWD
rerun=$repo/stage3/census/tsc-closure/rerun-2026-10-09
export GOPROXY='https://proxy.golang.org|direct'
export GOFLAGS=-p=2
timeout 120 git fetch --no-recurse-submodules origin f1502d130bc0b4440f29b93b76b7d18bff3f6a60 > "$scratch/fetch.log" 2>&1
GOFLAGS=-p=2 timeout 600 bash cloud/setup.sh > "$scratch/setup.log" 2>&1
source "${ADAMIC_TOOLS:-/workspace/adamic-tools}/env.sh"
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
timeout 600 bash stage3/apply.sh "$scratch/adapted" > "$scratch/apply.log" 2>&1
timeout 600 python3 stage3/census/tsc-closure/prepare.py "$repo" "$scratch/tools" > "$scratch/prepare.log" 2>&1
timeout 30 python3 "$rerun/compatibility.py" "$scratch/tools" "$scratch/fixture-runner" > "$scratch/compatibility.log" 2>&1
timeout 30 gofmt -w "$scratch/tools/overlay/internal_lower_latent_full.go"
(cd "$scratch/tools/tree"; GOWORK="$scratch/tools/go.work" timeout 300 go build -p 2 -overlay="$scratch/tools/overlay/overlay.json" -o "$scratch/tools/census" ./stage3/census/latent/tool) > "$scratch/build.log" 2>&1
timeout 90 node stage3/census/tsc-closure/closure.cjs "$scratch/adapted" src/tsc/tsc.ts "$scratch/closure.json" > "$scratch/graph.log" 2>&1
LATENT_ROOT_MANIFEST="$scratch/closure.json" LATENT_ASSERT_NO_OUTPUT=1 timeout 3600 "$scratch/tools/census" "$scratch/adapted/src/compiler" "$scratch/latent.jsonl" > "$scratch/latent.log" 2>&1
LATENT_ROOT_MANIFEST="$scratch/closure.json" LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 timeout 3600 "$scratch/tools/census" "$scratch/adapted/src/compiler" "$scratch/full.jsonl" > "$scratch/full.log" 2>&1
timeout 30 python3 - "$scratch" <<'PY'
import json,shutil,sys
from pathlib import Path
scratch=Path(sys.argv[1]); manifest=json.loads((scratch/'closure.json').read_text());root=Path(manifest['root']);view=scratch/'view';view.mkdir()
for file in manifest['files']:
    target=view/file['file'];target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(root/file['file'],target)
PY
timeout 90 node "$scratch/tools/tree/stage3/census/hidden/units.cjs" "$scratch/view" "$scratch/stock.json" > "$scratch/stock.log" 2>&1
timeout 90 python3 stage3/census/tsc-closure/measure.py "$scratch/closure.json" "$scratch/latent.jsonl" "$scratch/full.jsonl" "$scratch/stock.json" "$scratch/tools/tree/stage3/census/hidden/hidden.py" "$scratch/RESULT.json" > "$scratch/result.log" 2>&1
timeout 90 python3 "$scratch/fixture-runner/test_fixture.py" "$scratch/tools/census" "$scratch/tools/tree/stage3/census/hidden" "$scratch/fixture" > "$scratch/fixture.log" 2>&1
timeout 90 python3 "$rerun/audit.py" "$scratch/full.jsonl" "$scratch/stock.json" "$scratch/RESULT.json" > "$scratch/audit.log" 2>&1
