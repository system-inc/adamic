import os,sys,json,subprocess,hashlib,shutil,difflib
from pathlib import Path
import argparse
parser=argparse.ArgumentParser(description='Replay every adapter on both complete trees, continuing after refusals.')
for name in ['before_repo','after_repo','before_tree','after_tree','output']:parser.add_argument(name,type=Path)
args=parser.parse_args()
base=args.output.resolve();base.mkdir(parents=True,exist_ok=False)
hook=Path(__file__).with_name('replay-hook.cjs').resolve()
def snapshot(tree):
 return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
records=[]
for side,repo in [('before',args.before_repo.resolve()),('after',args.after_repo.resolve())]:
 source=(args.before_tree if side=='before' else args.after_tree).resolve(); tree=base/f'95h-replay-{side}'
 shutil.copytree(source,tree,ignore=shutil.ignore_patterns('.git','node_modules','built'))
 logdir=base/f'95h-replay-{side}-logs';logdir.mkdir()
 rows=[]
 for directory in sorted((repo/'stage3/adapt').iterdir()):
  if not directory.is_dir():continue
  name=directory.name;before=snapshot(tree);writes=logdir/(name+'.writes.json')
  env=dict(os.environ,NODE_PATH=str(args.after_repo.resolve()/'stage3/api/node_modules'),REPLAY_TREE=str(tree),REPLAY_LOG=str(writes),NODE_OPTIONS='--require='+str(hook))
  result=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
  (logdir/(name+'.stdout')).write_bytes(result.stdout);(logdir/(name+'.stderr')).write_bytes(result.stderr)
  after=snapshot(tree)
  changed=sorted(p for p in before.keys()|after.keys() if before.get(p)!=after.get(p))
  normalize=lambda text:text.replace(str(tree),'TREE').replace(str(repo),'REPO')
  row=dict(adapter=name,exit=result.returncode,filesWritten=sorted(item for line in writes.read_text().splitlines() for item in json.loads(line)) if writes.exists() else [],changedFiles=changed,stdout=normalize(result.stdout.decode()),stderr=normalize(result.stderr.decode()))
  rows.append(row);print(side,name,result.returncode,len(row['filesWritten']),len(changed),flush=True)
 (base/f'95h-replay-{side}.json').write_text(json.dumps(rows,indent=2)+'\n');records.append(rows)
a,b=records
text=''.join(difflib.unified_diff(json.dumps(a,indent=2).splitlines(True),json.dumps(b,indent=2).splitlines(True),fromfile='before',tofile='after'))
(base/'95h-replay.diff').write_text(text)
assert len(a)==len(b)
for x,y in zip(a,b):
 assert x['adapter']==y['adapter']
 if x['adapter']!='95-checker-axolotl-h':assert x==y, x['adapter']
 else:
  assert x['exit']!=0 and y['exit']!=0
  assert 'writer owner shape: createProgram' in y['stderr']
  assert 'writer owner shape' in x['stderr']
print('paired replay outcomes match except 95h; 95h reaches createProgram',flush=True)
