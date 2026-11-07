"""Run both censuses on one completed apply, with the bounded artifact probe alongside roots."""
import json,os,subprocess,sys,time
from pathlib import Path
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;core=repo/'stage3/census/core-sites-af42d3dc'
start=time.monotonic()
required=['setup-status.json','apply-status.json','entry-build-status.json']
while not all((scratch/n).exists() for n in required) or not (out/'CHECKPOINT.json').exists():
 for n in required:
  if (scratch/n).exists():assert json.loads((scratch/n).read_text())['exit_code']==0,(n,(scratch/n).read_text())
 if time.monotonic()-start>5400:raise RuntimeError('preparation has not completed within 90 minutes')
 time.sleep(2)
for n in required:assert json.loads((scratch/n).read_text())['exit_code']==0
for name,cmd in [
 ('core-runner',[sys.executable,str(core/'run-census.py')]),
 ('ast-map',['node',str(core/'ast-map.cjs')]),
 ('analysis',[sys.executable,str(core/'analyze.py')]),
 ('emitter-map',[sys.executable,str(core/'map-emitters.py')]),
 ('site-comparison',[sys.executable,str(core/'compare-sites.py')]),
 ('core-checkpoint',[sys.executable,str(out/'checkpoint.py'),'core census complete'])]:
 status=run(name,cmd,limit=3600)
 if status['exit_code']:sys.exit(1)
# The entry probe has its own 1800-second bound. Every child reports its actual exit code.
with (scratch/'entry-runner.log').open('wb') as log:
 entry=subprocess.Popen([sys.executable,str(out/'run-entry.py')],cwd=repo,stdout=log,stderr=log,start_new_session=True)
 (scratch/'entry-runner-pid.json').write_text(json.dumps({'pid':entry.pid})+'\n')
 try:
  result=run('groups',[sys.executable,str(out/'run-groups.py')],watch=list(out.glob('group-*.jsonl')),limit=7200)
  entry_rc=entry.wait();print('entry runner child_exit='+str(entry_rc),flush=True)
 finally:
  if entry.poll() is None:
   import signal
   os.killpg(entry.pid,signal.SIGTERM);entry.wait()
 if result['exit_code']:sys.exit(1)
print('measurements completed',flush=True)
