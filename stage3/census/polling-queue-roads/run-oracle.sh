#!/usr/bin/env bash
set -euo pipefail
cd /tmp/polling-queue-roads

test -f stage3/census/polling-queue-roads/evidence/sys.original.ts
python3 - <<'PY'
import shutil
shutil.copytree('/tmp/polling-adapted','/tmp/polling-oracle-tree-complete',symlinks=True,ignore=lambda directory,names:['.git'] if directory=='/tmp/polling-adapted' else [])
shutil.copyfile('stage3/census/polling-queue-roads/evidence/sys.original.ts','/tmp/polling-oracle-tree-complete/src/compiler/sys.ts')
PY
test -f stage3/census/polling-queue-roads/evidence/summary.json
export NODE_OPTIONS="--require=$PWD/stage3/census/polling-queue-roads/oracle-count.cjs"
export POLLING_COUNT_DIR=/tmp/polling-oracle-counts-complete
stage3/oracle/run.sh /tmp/polling-oracle-tree-complete /tmp/polling-oracle-results-complete
