import subprocess,time,json,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
while not (out/'apply.status.json').exists():time.sleep(1)
status=json.loads((out/'apply.status.json').read_text());print('apply observed own_child_exit='+str(status['exit']),flush=True)
if status['exit']:
 print('STOP: apply refused; no census will run',flush=True);raise SystemExit(status['exit'])
(out/'patch-set.md').write_bytes(Path('/workspace/scratch/fa-adapted/patch-set.md').read_bytes())
(out/'REPORT.md').write_text('# Stacked findAncestor census checkpoint\n\nTask 448a51k. Parent 1060360b; head 63c11488 (remote tip verified unchanged). Both required adaptation directories were copied from eeb87728. Apply succeeded on TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Compiler builds and censuses are pending. No regression verdict yet. Checkpoint refs are separate so the final branch can contain one commit directly on 1060360b without force-pushing.\n')
rc=subprocess.run([sys.executable,str(out/'checkpoint.py'),'apply'],cwd=repo).returncode
print('apply checkpoint wrapper_exit='+str(rc),flush=True)
if rc:raise SystemExit(rc)
while not (out/'head-focus-build.status.json').exists():time.sleep(1)
for side in ['parent','head']:
 for phase in ['overlay','entry-overlay','full-build','entry-build','focus-build']:
  status=json.loads((out/(side+'-'+phase+'.status.json')).read_text())
  if status['exit']:raise SystemExit(side+' '+phase+' failed '+str(status['exit']))
# Setup status is written by the launching shell once its own child exits.
while not Path('/tmp/fa-stacked-setup.status').exists():time.sleep(1)
assert Path('/tmp/fa-stacked-setup.status').read_text().strip()=='0'
(out/'setup.log').write_bytes(Path('/tmp/fa-stacked-setup.log').read_bytes())
rc=subprocess.run([sys.executable,str(out/'census.py')],cwd=repo).returncode
print('census controller own_child_exit='+str(rc),flush=True)
(out/'census-controller.status.json').write_text(json.dumps(dict(exit=rc))+'\n')
raise SystemExit(rc)
