import hashlib,json,time,tarfile,shutil
from pathlib import Path
s=Path('/workspace/scratch/core-extend');out=Path('/workspace/adamic/stage3/adapt/99-correctness-core-extend')
while True:
 summary=json.loads((s/'count-summary.json').read_text())
 for record in summary['pieces']:
  if record['failures']:continue
  i=record['index'];archive=out/f'count-piece-{i:02d}.tar.gz'
  if archive.exists():continue
  folder=s/f'piece-{i:02d}';(folder/'piece-summary.json').write_text(json.dumps(record,indent=2)+'\n')
  (folder/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(folder))+'\n' for p in sorted(folder.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
  with tarfile.open(archive,'w:gz') as tar:tar.add(folder,arcname=f'piece-{i:02d}')
  (out/f'count-piece-{i:02d}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
  print('packed',i,archive.stat().st_size,flush=True)
 shutil.copyfile(s/'count-summary.json',out/'COUNT-PROGRESS.json')
 if summary['complete']:break
 time.sleep(20)
