#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/cast-tags/env.sh
python3 - <<'PY'
from pathlib import Path
import time
while Path('/proc/2165').exists():time.sleep(5)
assert Path('/workspace/scratch/cast-tags/latent').exists()
PY
go build -overlay=/workspace/scratch/cast-tags/trace-overlay.json -o /workspace/scratch/cast-tags/trace-replay ./stage3/census/latent/replay/worker > /workspace/scratch/cast-tags/evidence/trace-build.log 2>&1
python3 - <<'PY'
from pathlib import Path
import time,subprocess,os,json
r=Path('/workspace/scratch/cast-tags');e=r/'evidence'
while not (e/'c-probe-owners.json').exists():time.sleep(5)
for side in ['main','c']:
 tree=r/(side+'-tree');source=tree/'src/compiler/checker.ts';lines=source.read_text().splitlines();line=next(i+1 for i,x in enumerate(lines) if 'source.parameterIndex !== (target as IdentifierTypePredicate)' in x);column=lines[line-1].index('target as')+1
 where=f'{source}:{line}:{column}'
 with (e/(side+'-checker-target-replay.json')).open('w') as out,(e/(side+'-checker-target-replay.log')).open('w') as log:
  p=subprocess.run([str(r/'trace-replay'),'-project',str(tree/'src/compiler'),'-where',where,'-kind','Refused','-reason',"a cast the runtime can't check"],env=dict(os.environ,LATENT_FULL='1',CAST_TAG_TRACE='1'),stdout=out,stderr=log)
 (e/(side+'-checker-target-replay-status.json')).write_text(json.dumps({'exit':p.returncode,'where':where})+'\n')
 print(side,p.returncode,where,flush=True)
PY
