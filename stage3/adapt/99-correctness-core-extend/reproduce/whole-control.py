import os,time,subprocess,json,signal
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');cwd=s/'main'
while not (s/'before/patch-set.md').exists():
 print('waiting for baseline apply',flush=True);time.sleep(20)
while any(line.strip().split(None,1)[1].startswith('python3 ') and '/stage3/apply.py --build-product /workspace/scratch/core-extend/' in line for line in subprocess.check_output(['ps','-eo','pid,args'],text=True).splitlines()[1:]):time.sleep(10)
start=time.monotonic();env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(s/'whole-control-tasks.jsonl'),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require='+str(repo/'stage3/oracle/profile-tasks.cjs'),PYTHONDONTWRITEBYTECODE='1')
with (s/'whole-control.log').open('w') as log:
 child=subprocess.Popen(['bash',str(cwd/'stage3/lane/run.sh'),str(s/'whole-control')],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=child.wait(timeout=1500)
 except subprocess.TimeoutExpired:
  os.killpg(child.pid,signal.SIGKILL);child.wait();code=124
(s/'whole-control-result.json').write_text(json.dumps({'exit':code,'seconds':time.monotonic()-start,'counted':False,'base':json.loads((s/'recut-base.json').read_text())['base']},indent=2)+'\n')
if code==0:(s/'count-workers.txt').write_text('4\n')
print('whole lane exit',code,flush=True)
