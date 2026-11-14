#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/changed-files-proof/env.sh
cd /workspace/scratch/changed-files-undefined-1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/changed-files-proof/overlay > /workspace/scratch/changed-files-proof/overlay.log 2>&1
python3 - <<'PY'
import json
from pathlib import Path
r=Path('/workspace/scratch/changed-files-proof/overlay')
p=r/'overlay.json'; data=json.loads(p.read_text())
for a,b in list(data['Replace'].items()):
 f=Path(b)
 if f.suffix=='.go':
  f.rename(str(f)+'.txt'); data['Replace'][a]=str(f)+'.txt'
p.write_text(json.dumps(data,indent=2)+'\n')
PY
gofmt -w /workspace/scratch/changed-files-proof/overlay/*.go.txt
go build -buildvcs=false -overlay=/workspace/scratch/changed-files-proof/overlay/overlay.json -o /workspace/scratch/changed-files-proof/latent-census ./stage3/census/latent/tool > /workspace/scratch/changed-files-proof/latent-build.log 2>&1
go build -buildvcs=false -o /workspace/scratch/changed-files-proof/adamic ./cmd/adamic > /workspace/scratch/changed-files-proof/adamic-build.log 2>&1
