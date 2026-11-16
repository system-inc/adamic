#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/cast-tags/env.sh
python3 - <<'PY'
from pathlib import Path
import time
while Path('/proc/936').exists() or Path('/proc/2165').exists():time.sleep(5)
assert Path('/workspace/scratch/cast-tags/main-tree/patch-set.md').exists(), 'baseline apply did not complete'
assert Path('/workspace/scratch/cast-tags/latent').exists(), 'census build did not complete'
PY
node /workspace/scratch/cast-tags/scout.cjs > /workspace/scratch/cast-tags/evidence/site-scout.jsonl
node /workspace/scratch/cast-tags/probe-c.cjs > /workspace/scratch/cast-tags/evidence/c-probe.log 2>&1
python3 /workspace/scratch/cast-tags/run-censuses.py
