import json,os,shutil,subprocess
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for label in ['before','after']:
 shutil.copytree(s/label,s/('replay-'+label),symlinks=True,ignore=shutil.ignore_patterns('node_modules','.git','built'))
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 def run(label):
  tree=s/('replay-'+label);r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env=env,capture_output=True,text=True)
  with (s/('replay-'+label+'.log')).open('a') as log:log.write(directory.name+'\n'+r.stdout+r.stderr)
  return label,{'exit':r.returncode,'stdout':r.stdout.replace(str(tree),'TREE'),'stderr':r.stderr.replace(str(tree),'TREE')}
 with ThreadPoolExecutor(max_workers=2) as pool:row={'adapter':directory.name,**dict(pool.map(run,['before','after']))}
 rows.append(row);(s/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
 if directory.name!='99-correctness-core-levenshtein':assert row['before']==row['after'],row
 else:
  assert row['before']['exit']==row['after']['exit']==0
  assert json.loads(row['before']['stdout'])=={'files':1,'edits':1,'javascriptChanging':True}
  assert json.loads(row['after']['stdout'])=={'files':0,'edits':0,'javascriptChanging':True}
 if directory.name=='30-indexed-reads':assert row['before']==row['after'],'STOP: candidate changes adapter 30'
assert next(r for r in rows if r['adapter']=='30-indexed-reads')['before']==next(r for r in rows if r['adapter']=='30-indexed-reads')['after']
print('PASS: only our adapter differs; 30 outcome identical (including its preexisting addRange refusal)',flush=True)

import hashlib
def manifest(label):
 root=s/('replay-'+label)/'src';return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
a,b=manifest('before'),manifest('after');assert a==b,'paired replay final source bytes differ'
(s/'replay-summary.json').write_text(json.dumps({'adapters':len(rows),'onlyOurOutcomeDiffers':True,'adapter30Identical':True,'finalSourcesByteIdentical':True,'sourceFiles':len(a)},indent=2)+'\n')
