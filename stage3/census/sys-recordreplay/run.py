import json,os,signal,subprocess,sys,time
from pathlib import Path
ROOT=Path('/workspace/adamic'); S=Path('/workspace/scratch/sys-walls')
def run(name,args,cwd=ROOT,env=None):
 log=S/(name+'.log'); start=time.monotonic(); reason=None
 with log.open('w') as stream:
  p=subprocess.Popen([str(a) for a in args],cwd=cwd,env=env,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
  last_size=-1; last_output=start
  try:
   while p.poll() is None:
    size=log.stat().st_size
    if size!=last_size:last_size=size;last_output=time.monotonic()
    if time.monotonic()-last_output>1800:
     reason='30 minutes without output';os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
     break
    time.sleep(1)
   code=p.wait()
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 record=dict(name=name,args=[str(a) for a in args],cwd=str(cwd),child_exit=code,seconds=round(time.monotonic()-start,3),stop_reason=reason)
 with (S/'commands.jsonl').open('a') as f:f.write(json.dumps(record)+'\n')
 print(json.dumps(record),flush=True)
 return code
if __name__=='__main__':sys.exit(run(sys.argv[1],sys.argv[2:]))
