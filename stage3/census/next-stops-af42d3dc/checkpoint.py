"""Push append-only checkpoints without changing the final branch or its parent."""
import json,subprocess,sys,time,fcntl
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/census/next-stops-af42d3dc';scratch=Path('/workspace/scratch/after-wave')
branch='refs/heads/typescript/next-stops-after-wave-checkpoint'
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
cmd(['add','stage3/census/next-stops-af42d3dc','stage3/census/core-sites-af42d3dc'])
tree=cmd(['write-tree'])
parent=json.loads(state.read_text())['sha'] if state.exists() else 'af42d3dc1261a3db536064c3dfbb5eaa7e8b6405'
sha=cmd(['commit-tree',tree,'-p',parent],input=('census checkpoint: '+sys.argv[1]+'\n').encode())
cmd(['push','origin',sha+':'+branch])
state.write_text(json.dumps({'sha':sha,'label':sys.argv[1],'time':time.time()},indent=2)+'\n')
(out/'CHECKPOINT.json').write_text(state.read_text())
