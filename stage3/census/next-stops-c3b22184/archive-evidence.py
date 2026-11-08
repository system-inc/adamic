import hashlib,json,tarfile
from pathlib import Path
out=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/c3b22184')
for p in scratch.glob('*-status.json'):(out/p.name).write_bytes(p.read_bytes())
saved=scratch/'ordinary';saved.mkdir(exist_ok=True)
for p in out.glob('group-*.jsonl'):(saved/p.name).write_bytes(p.read_bytes())
files={}
for p in sorted(scratch.rglob('*')):
 if saved not in p.parents and p.is_file() and (p.suffix in ('.log','.jsonl','.json') or p.name.endswith('.go.txt')) and not p.name.startswith(('finalize','archive')):
  files['scratch/'+str(p.relative_to(scratch))]=p
for p in sorted(saved.glob('group-*.jsonl')):files['ordinary/'+p.name]=p
# Checkpoint copies of these logs are represented once by the scratch originals.
for p in out.glob('*.log'):
 if p.name not in [f.name for f in files.values()]:files['report/'+p.name]=p
manifest={name:dict(bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for name,p in files.items()}
archive=out/'raw-streams.tar.gz'
with tarfile.open(archive,'w:gz',compresslevel=9) as tar:
 for name,p in files.items():tar.add(p,arcname=name,recursive=False)
(out/'archive-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'raw-streams.tar.gz.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  raw-streams.tar.gz\n')
for p in list(out.glob('*.log'))+list(out.glob('group-*.jsonl')):p.unlink()
print('archived_files='+str(len(files))+' archive_bytes='+str(archive.stat().st_size))
