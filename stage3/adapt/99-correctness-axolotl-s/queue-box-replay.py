import json,os,shutil,subprocess,sys
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
repo=Path(__file__).resolve().parents[3]
before,after,out=map(Path,sys.argv[1:]);out.mkdir(parents=True,exist_ok=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for label,tree in [('before',before),('after',after)]:
 shutil.copytree(tree,out/('replay-'+label),symlinks=True,ignore=shutil.ignore_patterns('node_modules'))
 (out/('replay-'+label)/'node_modules').symlink_to(before/'node_modules',target_is_directory=True)
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 def run(label):
  tree=(out/('replay-'+label)).resolve();r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env=env,capture_output=True,text=True)
  with (out/('replay-'+label+'.log')).open('a') as log:log.write(directory.name+'\n'+r.stdout+r.stderr)
  return label,{'exit':r.returncode,'stdout':r.stdout.replace(str(tree),'TREE'),'stderr':r.stderr.replace(str(tree),'TREE')}
 with ThreadPoolExecutor(max_workers=2) as pool:row={'adapter':directory.name,**dict(pool.map(run,['before','after']))}
 rows.append(row);(out/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
 if directory.name!='99-correctness-axolotl-s':assert row['before']==row['after'],row
 else:
  assert row['before']['exit']==row['after']['exit']==0
  assert json.loads(row['before']['stdout'])=={'files':1,'edits':1,'javascriptChanging':True}
  assert json.loads(row['after']['stdout'])=={'files':0,'edits':0,'javascriptChanging':True}
print('Only the 99s queue amendment differs',flush=True)
