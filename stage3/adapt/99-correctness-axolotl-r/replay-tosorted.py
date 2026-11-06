"""Paired ordered replay on independent copies of both complete source trees."""
import json,os,subprocess,shutil
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
repo=Path('/workspace/adamic'); scratch=Path('/workspace/scratch/tosorted/step2');raw=scratch/'raw'
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for side in ['before','after']:
 target=scratch/('replay-'+side)
 shutil.copytree(scratch/side,target,symlinks=True)
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 row={'adapter':directory.name}
 def replay(side):
  tree=scratch/('replay-'+side)
  r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env=env,capture_output=True,text=True)
  with (raw/('replay-'+side+'.log')).open('a') as log:log.write(directory.name+'\n'+r.stdout+r.stderr)
  # Node stack traces include scratch paths; normalize only the paired tree names.
  return side,{'exit':r.returncode,'stdout':r.stdout.replace(str(tree),'TREE'),'stderr':r.stderr.replace(str(tree),'TREE')}
 with ThreadPoolExecutor(max_workers=2) as pool:row.update(pool.map(replay,['before','after']))
 rows.append(row)
 (raw/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
for row in rows:
 if row['adapter']!='99-correctness-axolotl-r':assert row['before']==row['after'],row['adapter']
assert next(r for r in rows if r['adapter']=='99-correctness-axolotl-r')['before']['exit']==1
print('All non-99r outcomes identical',flush=True)
