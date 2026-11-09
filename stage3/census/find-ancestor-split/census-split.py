import os,sys,subprocess,json,time,signal,hashlib
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
baseenv={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
baseenv.update(LATENT_ASSERT_NO_OUTPUT='1',GOMEMLIMIT='5GiB',GOGC='70',ADAMIC_TYPESCRIPT_SOURCE='/workspace/scratch/fa-typescript')
active={}
checked_headers=set()
with (out/'baseline-parent-entry-run.jsonl').open() as f: expected_header=json.loads(next(f))
def checkpoint(event):
 rc=subprocess.run([sys.executable,str(out/'checkpoint.py'),event],cwd=repo).returncode
 print(event+' checkpoint_wrapper_exit='+str(rc),flush=True)
 if rc:raise RuntimeError('checkpoint failed: '+str(rc))
try:
 for side in ['a','b']:
  name=side+'-entry-run';env=dict(baseenv,LATENT_TIMES=str(out/(name+'-units.jsonl')))
  active[name]=subprocess.Popen([sys.executable,str(out/'run.py'),name,str(repo),'/workspace/scratch/fa-'+side+'-entry-bin','/workspace/scratch/fa-adapted/src/tsc/tsc.ts',str(out/(name+'.jsonl'))],env=env)
 last=0
 while active:
  for name in list(active):
   stream=out/(name+'.jsonl')
   if name not in checked_headers and stream.exists():
    with stream.open() as f: line=f.readline()
    if line.endswith('\n'):
     actual=json.loads(line)
     if actual!=expected_header:raise RuntimeError(name+' header differs from reused parent')
     checked_headers.add(name);print(name+' header matches reused parent',flush=True)
  for name,p in list(active.items()):
   if p.poll() is not None:
    rc=p.wait();del active[name]
    print(name+' wrapper_exit='+str(rc),flush=True)
    checkpoint(name)
    if rc:raise RuntimeError(name+' failed '+str(rc))
  if time.monotonic()-last>=60:
   print('entry active '+','.join(active),flush=True);last=time.monotonic()
  time.sleep(1)
finally:
 for name,p in active.items():
  pidfile=out/(name+'.pid.json')
  if pidfile.exists():
   try:os.killpg(json.loads(pidfile.read_text())['pid'],signal.SIGTERM)
   except ProcessLookupError:pass
  try:p.wait(timeout=15)
  except subprocess.TimeoutExpired:
   if pidfile.exists():
    try:os.killpg(json.loads(pidfile.read_text())['pid'],signal.SIGKILL)
    except ProcessLookupError:pass
   p.terminate();p.wait()
print('census complete child_exit=0',flush=True)
