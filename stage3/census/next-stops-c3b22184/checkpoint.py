"""Push append-only checkpoints without changing the final branch or its parent."""
import json,subprocess,sys,time,fcntl
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/census/next-stops-c3b22184';scratch=Path('/workspace/scratch/c3b22184')
branch='refs/heads/typescript/next-stops-c3b22184-checkpoint'
state=scratch/'checkpoint-state.json'
lock=(scratch/'checkpoint.lock').open('a')
fcntl.flock(lock,fcntl.LOCK_EX)
def cmd(args,**kw):
 p=subprocess.run(['git',*args],cwd=repo,capture_output=True,**kw)
 with (scratch/'checkpoint-git.log').open('ab') as log:log.write(p.stdout+p.stderr)
 print('git '+args[0]+' child_exit='+str(p.returncode),flush=True);p.check_returncode();return p.stdout.decode().strip()
for name in ['setup.log','npm-ci.log','source-clone.log','apply.log','apply-status.json']:
 p=scratch/name
 if p.exists(): (out/name).write_bytes(p.read_bytes())
for p in scratch.glob('*-status.json'):(out/p.name).write_bytes(p.read_bytes())
cmd(['add','stage3/census/next-stops-c3b22184'])
tree=cmd(['write-tree'])
parent=json.loads(state.read_text())['sha'] if state.exists() else 'c3b22184745d809cb281a37ef1989c895596223f'
sha=cmd(['commit-tree',tree,'-p',parent],input=('census checkpoint: '+sys.argv[1]+'\n').encode())
cmd(['push','origin',sha+':'+branch])
state.write_text(json.dumps({'sha':sha,'label':sys.argv[1],'time':time.time()},indent=2)+'\n')
(out/'CHECKPOINT.json').write_text(state.read_text())
