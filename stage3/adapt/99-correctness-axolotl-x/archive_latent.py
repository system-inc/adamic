#!/usr/bin/env python3
"""Validate and retain one completed CW051 measurement piece."""
import gzip
import hashlib
import json
from pathlib import Path
import shutil
import sys
side,index=sys.argv[1:];index=int(index)
p=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch');root='/workspace/bind-'+side+'/'
plan=json.loads((scratch/'bind-latent-plans'/side/(str(index)+'.json')).read_text())
metrics=json.loads((scratch/f'bind-latent-{side}-{index}/metrics.json').read_text());assert metrics['exit']==0 and metrics['seconds']<1500
records=scratch/f'bind-latent-{side}-{index}-records.jsonl'
rows=[json.loads(s) for s in records.read_text().splitlines()]
assert rows[0]['checker_rejected'] is False, rows[0]
files=[r for r in rows if 'units' in r];assert len(files)==1
units=files[0]['units'];assert len(units)==1
unit=units[0];norm=lambda s:s.removeprefix(root)
assert norm(unit['where'])==plan['expected_units'][0]
findings=[f for f in files[0]['findings'] if f.get('phase')=='lowering']
assert all(norm(f['unit'])==plan['expected_units'][0] for f in findings)
assert len(findings)<=1
stop={k:findings[0].get(k) for k in ['kind','where','reason']} if findings else dict(kind=unit['status'],where=unit['where'],reason='no lowering finding')
stop['where']=norm(stop['where'])
summary=dict(side=side,index=index,reference=plan['reference'],name=plan['name'],actual=norm(unit['where']),unit_status=unit['status'],checker_rejected=rows[0]['checker_rejected'],stop=stop,metrics=metrics,binary_sha256=hashlib.sha256((scratch/'bind-bin/latent').read_bytes()).hexdigest())
dest=p/'latent'/side/str(index);dest.mkdir(parents=True,exist_ok=True)
(dest/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
(dest/'selection.json').write_text(json.dumps(plan,indent=2)+'\n')
for source,name in [(records,'records.jsonl.gz'),(scratch/f'bind-latent-{side}-{index}-times.jsonl','times.jsonl.gz'),(scratch/f'bind-latent-{side}-{index}/process.log','process.log.gz')]:
 if source.exists():
  with gzip.open(dest/name,'wb') as out:out.write(source.read_bytes())
print(json.dumps(summary),flush=True)
