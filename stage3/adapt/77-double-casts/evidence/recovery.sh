#!/usr/bin/env bash
# Reproduction of the observed timeout recovery, on the already applied tree.
set -u
source /workspace/adamic-tools/env.sh
mkdir /tmp/adapt77-final
ln -s /tmp/adapt77-recovered/adapted-tree /tmp/adapt77-final/adapted-tree
cp /tmp/adapt77-recovered/apply.log /tmp/adapt77-final/apply.log
cp /tmp/adapt77-recovered/patch-set.md /tmp/adapt77-final/patch-set.md
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/oracle/run.sh /tmp/adapt77-final/adapted-tree /tmp/adapt77-final/oracle --workers=4 > /tmp/adapt77-final/oracle.log 2>&1
oracle_status=$?
python3 - "$oracle_status" <<'PY'
import json,sys
from pathlib import Path
p=Path('/tmp/adapt77-final')
x=json.loads(Path('/tmp/adapt77-recovered/execution.json').read_text())
x['oracle_exit']=int(sys.argv[1])
x['recovery']='Retained exact successful apply output; reran full oracle at four workers after an upstream 40-second hook timeout at eight workers.'
x['wall_seconds']=json.loads((p/'oracle/report.json').read_text())['wall_seconds']
(p/'execution.json').write_text(json.dumps(x,indent=2)+'\n')
PY
python3 stage3/lane/check.py /tmp/adapt77-final > /tmp/adapt77-final-lane.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules NODE_OPTIONS=--max-old-space-size=1400 node stage3/adapt/77-double-casts/proof.cjs /tmp/adapt77-before /tmp/adapt77-final /tmp/adapt77-final-proof > /tmp/adapt77-final-proof.log 2>&1
