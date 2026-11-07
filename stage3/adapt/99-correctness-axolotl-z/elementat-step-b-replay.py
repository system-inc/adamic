import json,os,shutil,subprocess
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/elementat-b/step-b');raw=s/'raw';before=Path('/workspace/scratch/elementat-b/whole-control/adapted-tree');after=s/'after';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for label,tree in [('before',before),('after',after)]:shutil.copytree(tree,s/('replay-'+label),symlinks=True,ignore=shutil.ignore_patterns('node_modules'));(s/('replay-'+label)/'node_modules').symlink_to(s/'packages/node_modules',target_is_directory=True)
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 def run(label):
  tree=s/('replay-'+label);r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env=env,capture_output=True,text=True)
  with (raw/('replay-'+label+'.log')).open('a') as log:log.write(directory.name+'\n'+r.stdout+r.stderr)
  return label,{'exit':r.returncode,'stdout':r.stdout.replace(str(tree),'TREE'),'stderr':r.stderr.replace(str(tree),'TREE')}
 with ThreadPoolExecutor(max_workers=2) as pool:row={'adapter':directory.name,**dict(pool.map(run,['before','after']))}
 rows.append(row);(raw/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
 if directory.name!='99-correctness-axolotl-z':assert row['before']==row['after'],row
 else:
  assert row['before']['exit']==row['after']['exit']==0
  assert json.loads(row['before']['stdout'])=={'files':1,'edits':1,'javascriptChanging':True}
  assert json.loads(row['after']['stdout'])=={'files':0,'edits':0,'javascriptChanging':True}
print('Only 99z differs',flush=True)
