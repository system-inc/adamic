#!/usr/bin/env python3
from pathlib import Path
import hashlib,json,tarfile
here=Path(__file__).resolve().parent
results=Path('/workspace/scratch/polling-shards')
rows=json.loads((here/'shard-exits.json').read_text())
assert len(rows)==398 and all(row['exit']==0 for row in rows)
manifest={}
with tarfile.open(here/'shard-raw-evidence.tar.gz','w:gz') as archive:
 for row in rows:
  for name in ['report.json','verdict.json','observations.json','failures.md','oracle/tests.log','oracle/baseline.diff']:
   source=results/row['name']/name
   assert source.is_file(),source
   relative=row['name']+'/'+name
   manifest[relative]=hashlib.sha256(source.read_bytes()).hexdigest()
   archive.add(source,arcname=relative,recursive=False)
(here/'shard-archive-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(json.dumps(dict(files=len(manifest),archive_sha256=hashlib.sha256((here/'shard-raw-evidence.tar.gz').read_bytes()).hexdigest())))
