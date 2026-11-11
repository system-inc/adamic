import json,tarfile
from pathlib import Path
raw=Path(__file__).parent;root=Path('/workspace/scratch/meta-spread-lane-final')
summary=json.loads((raw/'lane-summary.json').read_text());assert summary['passed'] and summary['shards']==398
with tarfile.open(raw/'lane-evidence.tar.gz','w:gz') as archive:
 for p in sorted(root.rglob('*')):
  rel=p.relative_to(root)
  if any(part in ('adapted-tree','caches') for part in rel.parts):continue
  if p.is_file() and not p.is_symlink():archive.add(p,arcname=str(rel),recursive=False)
Path('/workspace/adamic/stage3/adapt/47-host-errors/meta-spread-lane.json').write_text(json.dumps(summary,indent=2)+'\n')
print('all 398 shard reports, observations, logs and diffs archived; private views excluded')
