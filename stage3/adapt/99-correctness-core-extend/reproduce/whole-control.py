import os,time,subprocess,json,signal
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');cwd=s/'control-main'
rows=subprocess.check_output(['ps','-eo','pid,args'],text=True).splitlines();pids=[int(r.strip().split(None,1)[0]) for r in rows if r.strip().split(None,1)[1:] == ['python3 /workspace/scratch/core-extend/count-pieces.py']]
assert len(pids)==1,pids;pid=pids[0];os.kill(pid,signal.SIGSTOP);start=time.monotonic();(s/'count-pause.json').write_text(json.dumps({'pid':pid,'reason':'whole stage3/lane/run.sh control; active child shards may finish','started':time.time()})+'\n')
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(s/'whole-control-tasks.jsonl'),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require='+str(repo/'stage3/oracle/profile-tasks.cjs'))
try:
 with (s/'whole-control.log').open('w') as log:
  child=subprocess.Popen(['bash',str(cwd/'stage3/lane/run.sh'),str(s/'whole-control')],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=1400)
  except subprocess.TimeoutExpired:
   os.killpg(child.pid,signal.SIGKILL);child.wait();code=124
finally:
 os.kill(pid,signal.SIGCONT)
(s/'whole-control-result.json').write_text(json.dumps({'exit':code,'seconds':time.monotonic()-start,'countPauseIncludedInPieceTiming':True,'counted':False},indent=2)+'\n');print(code,flush=True)
