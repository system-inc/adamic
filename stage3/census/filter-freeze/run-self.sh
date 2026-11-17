#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/filter-freeze/env.sh
cd /workspace/scratch/filter-freeze/adapted
FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/build-entry-counts NODE_OPTIONS=--require=/workspace/scratch/filter-freeze/repo/stage3/census/filter-freeze/observe.cjs npm run build > /workspace/scratch/filter-freeze/build-self.log 2>&1
export FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/self-entry-counts
export NODE_OPTIONS='--require=/workspace/scratch/filter-freeze/repo/stage3/census/filter-freeze/observe.cjs'
set +e
node built/local/tsc.js -b src/compiler src/tsc --force --pretty false > /workspace/scratch/filter-freeze/self.log 2>&1
status=$?
python3 - "$status" <<'PY'
import json,sys
from pathlib import Path
Path('/workspace/scratch/filter-freeze/self-execution.json').write_text(json.dumps({'command':['node','built/local/tsc.js','-b','src/compiler','src/tsc','--force','--pretty','false'],'exit':int(sys.argv[1])},indent=2)+'\n')
PY
exit "$status"
