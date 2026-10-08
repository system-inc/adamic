#!/usr/bin/env bash
set -euo pipefail
: "${ADAMIC_TYPESCRIPT_SOURCE:?set this to the pinned TypeScript 6.0.3 checkout}"
task_root=$(git rev-parse --show-toplevel)
cd "$task_root"
task_scratch=$(mktemp -d "${ADAMIC_LINT_WORK_DIR:-/tmp}/wave13-lint-verify.XXXXXX")
python3 - "$task_root" "$task_scratch" <<'PY'
import json, pathlib, sys
root, scratch = map(pathlib.Path, sys.argv[1:])
source = root / 'stage1/cohere/lint/rules/nexus-consistency-no-hand-rolled-delay/evidence/native_mutants_test.go.txt'
(scratch / 'overlay.json').write_text(json.dumps({'Replace': {str(root / 'stage1/cohere/lint/wave13_worker_test.go'): str(source)}}))
PY
export XDG_CACHE_HOME="$task_scratch/cache"
export GOCACHE="$task_scratch/go-cache"
export ADAMIC_LINT_BENCH=1
export ADAMIC_LINT_PROFILE_DIR="$task_scratch/profile"
export ADAMIC_LINT_PROFILE_SNAPSHOTS="$task_scratch/profile"
go test -overlay "$task_scratch/overlay.json" ./stage1/cohere/lint -count=1 -json -timeout=60m > "$task_scratch/lint.jsonl" 2>&1
printf 'Lint log: %s/lint.jsonl\n' "$task_scratch"
