#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/core-sites-main
while [ ! -f /workspace/scratch/b5w3ycg/setup-result.txt ]; do sleep 2; done
[ "$(cut -d' ' -f1 /workspace/scratch/b5w3ycg/setup-result.txt)" = 0 ]
python3 stage3/census/latent/make_overlay.py /workspace/core-sites-main /workspace/scratch/b5w3ycg/overlay > /workspace/scratch/b5w3ycg/overlay-retry.log 2>&1
python3 - <<'PY'
import pathlib,json
p=pathlib.Path('/workspace/scratch/b5w3ycg/overlay/overlay.json')
data=json.loads(p.read_text())
for k,v in data['Replace'].items():
 path=pathlib.Path(v)
 if path.suffix=='.go':
  new=path.with_suffix('.go.txt');path.rename(new);data['Replace'][k]=str(new)
p.write_text(json.dumps(data,indent=2)+'\n')
PY
gofmt -w /workspace/scratch/b5w3ycg/overlay/*.go.txt > /workspace/scratch/b5w3ycg/gofmt.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/b5w3ycg/overlay/overlay.json -o /workspace/scratch/b5w3ycg/latent ./stage3/census/latent/tool > /workspace/scratch/b5w3ycg/build.log 2>&1
date -u +%FT%TZ > /workspace/scratch/b5w3ycg/build-complete.txt
