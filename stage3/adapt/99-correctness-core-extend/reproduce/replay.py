import os,time,subprocess,signal,json,shutil,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');rows=json.loads((s/'paired-replay.json').read_text()) if (s/'paired-replay.json').exists() else []
while not all((s/n/'patch-set.md').exists() for n in ['before','after']):
 print('waiting for complete applies',flush=True);time.sleep(30)
for side in ['before','after']:
 d=s/('replay-'+side)
 if d.exists():continue
 shutil.copytree(s/side,d,symlinks=True,ignore=shutil.ignore_patterns('.git','node_modules','built','tests'))
 (d/'node_modules').symlink_to(s/'before/node_modules',target_is_directory=True)
def snap(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for directory in sorted((repo/'stage3/adapt').iterdir()):
 if not directory.is_dir() or directory.name in [r['adapter'] for r in rows]:continue
 pair={'adapter':directory.name,'sides':{}}
 for side in ['before','after']:
  d=s/('replay-'+side);before=snap(d);begin=time.monotonic()
  with (s/('replay-'+side+'-'+directory.name+'.log')).open('w') as log:
   p=subprocess.Popen(['node',str(directory/'adapt.cjs'),str(d)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
   try:code=p.wait(timeout=1400)
   finally:
    try:os.killpg(p.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    p.wait()
  after=snap(d);pair['sides'][side]={'exit':code,'changed':[n for n in after if before.get(n)!=after[n]],'seconds':time.monotonic()-begin}
 rows.append(pair);(s/'paired-replay.json').write_text(json.dumps(rows,indent=2)+'\n');print(directory.name,pair['sides'],flush=True)
 assert pair['sides']['before']['exit']==pair['sides']['after']['exit'],pair
 expected=['src/compiler/core.ts'] if directory.name=='99-correctness-core-extend' else []
 assert pair['sides']['before']['changed']==expected,pair
 assert not pair['sides']['after']['changed'],pair
 if directory.name!='99-correctness-core-extend':
  logs=[(s/('replay-'+side+'-'+directory.name+'.log')).read_text().replace(str(s/('replay-'+side)),'<tree>') for side in ['before','after']]
  assert logs[0]==logs[1],pair
assert snap(s/'replay-before')==snap(s/'replay-after')==snap(s/'after')
