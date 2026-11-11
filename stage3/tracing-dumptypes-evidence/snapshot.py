from pathlib import Path
import shutil,gzip,json,hashlib
S=Path(__file__).resolve().parent;D=Path('/workspace/scratch/dp24r5m-evidence-branch/stage3/tracing-dumptypes-evidence');
if D.exists():shutil.rmtree(D)
D.mkdir(parents=True,exist_ok=True)
for f in S.iterdir():
 if f.is_file() and f.suffix in ['.py','.cjs','.json','.log','.txt','.ts','.patch','.md']:
  if f.stat().st_size>1000000:
   with f.open('rb') as src,gzip.open(D/(f.name+'.gz'),'wb') as dst:shutil.copyfileobj(src,dst)
  else:shutil.copyfile(f,D/f.name)
for name in ['stock-counts','stock-candidate-counts','stock-late-counts','project-counts','lane-counts','attempt1-lane-counts','overlay','flags-probe','in-probe']:
 p=S/name
 if not p.exists():continue
 for f in p.iterdir():
  if not f.is_file():continue
  target=D/name/(f.name+'.txt' if f.suffix=='.go' else f.name);target.parent.mkdir(parents=True,exist_ok=True)
  if f.suffix=='.jsonl':
   with f.open('rb') as src,gzip.open(str(target)+'.gz','wb') as dst:shutil.copyfileobj(src,dst)
  else:shutil.copyfile(f,target)
for p in S.glob('adapted-project-trace-*'):
 f=p/'types.json'
 if f.is_file():
  target=D/p.name/'types.json.gz';target.parent.mkdir(parents=True,exist_ok=True)
  with f.open('rb') as src,gzip.open(target,'wb') as dst:shutil.copyfileobj(src,dst)
manifest=[dict(path=str(f.relative_to(D)),bytes=f.stat().st_size,sha256=hashlib.sha256(f.read_bytes()).hexdigest()) for f in sorted(D.rglob('*')) if f.is_file() and f.name!='manifest.json']
(D/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(len(manifest),'evidence files')
