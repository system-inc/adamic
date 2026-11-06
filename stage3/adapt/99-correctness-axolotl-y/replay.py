import os, subprocess, json, shutil
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
repo=Path('/workspace/adamic'); out=repo/'stage3/adapt/99-correctness-axolotl-y'
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 row={'adapter':directory.name}
 def replay(side):
  with (out/f'replay-{side}.log').open('a') as log:
   r=subprocess.run(['node',str(directory/'adapt.cjs'),'/workspace/y-replay-'+side],env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
   log.write(directory.name+'\n'+r.stdout+r.stderr)
  return side,{'exit':r.returncode,'stdout':r.stdout.replace('/workspace/y-replay-'+side,'TREE'),'stderr':r.stderr.replace('/workspace/y-replay-'+side,'TREE')}
 with ThreadPoolExecutor(max_workers=2) as pool:
  row.update(pool.map(replay,['before','after']))
 rows.append(row)
 (out/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
for row in rows:
 if row['adapter']=='99-correctness-axolotl-y': continue
 assert row['before']==row['after'],row['adapter']
print('All non-99y outcomes identical')
