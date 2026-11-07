#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
while [ ! -f /workspace/scratch/core-compare/api-install.log ]; do sleep 2; done
python3 stage3/census/latent/make_overlay.py /workspace/adamic /workspace/scratch/core-compare/overlay > /workspace/scratch/core-compare/overlay.log 2>&1
python3 - <<'PY'
import json,pathlib
p=pathlib.Path('/workspace/scratch/core-compare/overlay/overlay.json');data=json.loads(p.read_text())
for k,v in data['Replace'].items():
 path=pathlib.Path(v)
 if path.suffix=='.go':
  new=path.with_suffix('.go.txt');path.rename(new);data['Replace'][k]=str(new)
p.write_text(json.dumps(data,indent=2)+'\n')
PY
gofmt -w /workspace/scratch/core-compare/overlay/*.go.txt > /workspace/scratch/core-compare/gofmt.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/core-compare/overlay/overlay.json -o /workspace/scratch/core-compare/latent ./stage3/census/latent/tool > /workspace/scratch/core-compare/latent-build.log 2>&1
go build -o /workspace/scratch/core-compare/adamic ./cmd/adamic > /workspace/scratch/core-compare/adamic-build.log 2>&1
printf '0\n' > /workspace/scratch/core-compare/tools.exit
