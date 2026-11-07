import json,os,shutil,subprocess,signal,hashlib,time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/core-compare');raw=s/'replay';raw.mkdir();env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for label in ['before','after']:shutil.copytree(s/label,s/('replay-'+label),symlinks=True,ignore=shutil.ignore_patterns('node_modules'));(s/('replay-'+label)/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 def run(label):
  tree=s/('replay-'+label);log=raw/(directory.name+'-'+label+'.log')
  with log.open('w') as stream:
   p=subprocess.Popen(['node',str(directory/'adapt.cjs'),str(tree)],env=env,stdout=stream,stderr=stream,start_new_session=True)
   try:code=p.wait(timeout=1800)
   finally:
    try:os.killpg(p.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    p.wait()
  return label,{'exit':code,'output':log.read_text().replace(str(tree),'TREE')}
 with ThreadPoolExecutor(max_workers=2) as pool:row={'adapter':directory.name,**dict(pool.map(run,['before','after']))}
 rows.append(row);(raw/'replay.json').write_text(json.dumps(rows,indent=2)+'\n');print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
 if directory.name!='99-correctness-core-compare':assert row['before']==row['after'],row
 else:
  assert row['before']['exit']==row['after']['exit']==0
  assert json.loads(row['before']['output'])=={'files':1,'edits':1,'javascriptChanging':True}
  assert json.loads(row['after']['output'])=={'files':0,'edits':0,'javascriptChanging':True}
print('Only compare differs',flush=True)
