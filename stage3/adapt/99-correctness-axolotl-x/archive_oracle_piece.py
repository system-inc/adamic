#!/usr/bin/env python3
"""Validate and retain one completed standard-lane shard chunk."""
import gzip
import json
from pathlib import Path
import shutil
import sys
results=Path(sys.argv[1]).resolve();index=int(sys.argv[2]);p=Path(__file__).resolve().parent
piece=results/('piece-'+str(index));record=json.loads((piece/'results.json').read_text());assert not record['failed'] and record['seconds']<1500
out=p/'oracle/pieces'/str(index);out.mkdir(parents=True,exist_ok=True)
for item in record['units']:
    assert item['exit']==0 and item['report']['status']=='pass'
    source=results/'units'/item['name'];dest=p/'oracle/shards'/item['name']
    dest.mkdir(parents=True,exist_ok=True)
    for f in source.rglob('*'):
        if not f.is_file() or 'adapted-tree' in f.relative_to(source).parts:continue
        target=dest/f.relative_to(source);target.parent.mkdir(parents=True,exist_ok=True)
        if f.suffix in ['.log','.jsonl'] or f.name=='observations.json':
            with gzip.open(str(target)+'.gz','wb') as archive:archive.write(f.read_bytes())
        else:shutil.copyfile(f,target)
with gzip.open(out/'results.json.gz','wb') as archive:archive.write((piece/'results.json').read_bytes())
metrics=Path('/workspace/scratch/bind-oracle-shards-'+str(index)+'/metrics.json')
assert json.loads(metrics.read_text())['exit']==0
shutil.copyfile(metrics,out/'metrics.json')
with gzip.open(out/'process.log.gz','wb') as archive:archive.write(metrics.with_name('process.log').read_bytes())
print(json.dumps(dict(piece=index,units=len(record['units']),seconds=record['seconds'],all_passed=True)))
