import json,hashlib,tarfile,subprocess,os
from pathlib import Path
from runner import out,raw,repo,run
for side in ['base','head']:
 for kind in ['census','latent']:
  label=side+('-binary.sha256' if kind=='census' else '-latent-binary.sha256')
  assert hashlib.sha256((raw/(side+'-'+kind)).read_bytes()).hexdigest()==(out/label).read_text().strip()
# Completed census invocations must be reaped and their original process groups absent.
tree=Path('/workspace/require-row-adapted')
manifest={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
assert manifest==json.loads((out/'adapted-source-manifest.json').read_text()),'adapted source changed during measurements'
cleanup=[]
for p in sorted(raw.glob('*-status.json')):
 s=json.loads(p.read_text())
 if '-root-' not in p.name and p.name not in ['base-core-status.json','head-core-status.json']:continue
 assert s['exit_code']==0 and s['process_group_cleaned']
 pid=s.get('pid');alive=False
 if pid:
  try:os.killpg(pid,0);alive=True
  except ProcessLookupError:pass
 assert not alive,(p.name,pid)
 cleanup.append(dict(run=p.name,pid=pid,process_group_exists=alive))
assert len(cleanup)==160,len(cleanup)
(out/'cleanup.json').write_text(json.dumps(cleanup,indent=2)+'\n')
run('render-report',['python3',str(out/'report.py')])
# Archive logs, exact streams and overlays, excluding the large reproducible binaries.
files=[p for p in sorted(raw.rglob('*')) if p.is_file() and p.name not in ['base-census','head-census','base-latent','head-latent']]
assert not list(raw.rglob('*.go'))
(out/'raw-member-sha256.json').write_text(json.dumps({str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in files},indent=2)+'\n')
archive=out/'raw-streams.tar.gz'
with tarfile.open(archive,'w:gz') as tar:
 for p in files:tar.add(p,arcname=str(p.relative_to(out)))
(out/'raw-streams.tar.gz.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  raw-streams.tar.gz\n')
# Raw members remain available locally and in checkpoint history, with one packed copy in the final commit.
(out/'.gitignore').write_text('__pycache__/\nraw/\n')
print('archive bytes',archive.stat().st_size)
