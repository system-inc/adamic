from pathlib import Path
import gzip,shutil
src=Path(__file__).resolve().parent;dest=Path('/workspace/scratch/single-watcher-evidence-branch/stage3/census/single-watcher-forwarder')
allowed={'.py','.cjs','.md','.txt','.json','.jsonl','.log','.sh','.ts','.js','.diff'}
def copy(p,target):
 target.parent.mkdir(parents=True,exist_ok=True)
 if p.suffix in ['.log','.jsonl']:
  with gzip.open(str(target)+'.gz','wb') as f:f.write(p.read_bytes())
 else:shutil.copyfile(p,target)
for p in src.rglob('*'):
 if p.is_file() and p.suffix in allowed:copy(p,dest/p.relative_to(src))
lane=Path('/workspace/scratch/single-watcher-lane')
if (lane/'report.json').exists():
 for p in lane.iterdir():
  if p.is_file():copy(p,dest/'lane'/p.name)
 oracle=lane/'oracle'
 for p in oracle.iterdir():
  if p.is_file():copy(p,dest/'lane/oracle'/p.name)
