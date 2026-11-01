#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
free -m > /workspace/scratch/99e-free-before.log
set +e
bash stage3/oracle/run.sh /workspace/scratch/99e-before /workspace/scratch/99e-oracle-before --workers 4 --limit-seconds 1800 > /workspace/scratch/99e-oracle-before.log 2>&1
result=$?
set -e
[ "$result" -le 1 ]
python3 - <<'PY'
import json
p=json.load(open('/workspace/scratch/99e-oracle-before/report.json'))
assert p['counts']==dict(passing=106366,failing=1,pending=0),p['counts']
assert p['baseline_diffs']==['api/typescript.d.ts'],p['baseline_diffs']
PY
free -m > /workspace/scratch/99e-free-after.log
set +e
bash stage3/oracle/run.sh /workspace/scratch/99e-after /workspace/scratch/99e-oracle-after --workers 4 --limit-seconds 1800 > /workspace/scratch/99e-oracle-after.log 2>&1
result=$?
set -e
[ "$result" -le 1 ]
python3 - <<'PY'
from pathlib import Path
import json
b=Path('/workspace/scratch/99e-oracle-before');a=Path('/workspace/scratch/99e-oracle-after')
x=json.loads((b/'report.json').read_text());y=json.loads((a/'report.json').read_text())
assert x['counts']==y['counts'],(x['counts'],y['counts'])
assert (b/'baseline.diff').read_bytes()==(a/'baseline.diff').read_bytes()
print('identical baseline bytes and counts')
PY
