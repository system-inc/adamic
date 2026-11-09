#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
python3 stage3/census/latent/make_overlay.py /workspace/adamic /workspace/scratch/levenshtein/overlay > /workspace/scratch/levenshtein/overlay.log 2>&1
python3 - <<'PY'
from pathlib import Path
import json
p=Path('/workspace/scratch/levenshtein/overlay/overlay.json');data=json.loads(p.read_text())
for k,v in data['Replace'].items():
 path=Path(v)
 if path.suffix=='.go':
  new=path.with_suffix('.go.txt');path.rename(new);data['Replace'][k]=str(new)
p.write_text(json.dumps(data,indent=2)+'\n')
PY
gofmt -w /workspace/scratch/levenshtein/overlay/*.go.txt > /workspace/scratch/levenshtein/gofmt.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/levenshtein/overlay/overlay.json -o /workspace/scratch/levenshtein/latent ./stage3/census/latent/tool > /workspace/scratch/levenshtein/census-build.log 2>&1
go build -o /workspace/scratch/levenshtein/adamic ./cmd/adamic > /workspace/scratch/levenshtein/adamic-build.log 2>&1
