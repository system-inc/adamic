from pathlib import Path
import os,subprocess,signal,time,json,sys
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/getfunctionname');raw=s/'raw'
def run(name,args,env=None,limit=1490,cwd=repo):
 start=time.monotonic();last=start;size=-1
 with (raw/(name+'.log')).open('w') as log:
  p=subprocess.Popen([str(a) for a in args],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  reason=None
  try:
   while p.poll() is None:
    current=(raw/(name+'.log')).stat().st_size
    if current!=size:size=current;last=time.monotonic()
    if time.monotonic()-start>limit or time.monotonic()-last>1800:
     reason='time bound' if time.monotonic()-start>limit else '30 minutes without output';os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
     break
    time.sleep(1)
  finally:
   if p.poll() is None:os.killpg(p.pid,signal.SIGKILL);p.wait()
 result={'command':[str(a) for a in args],'cwd':str(cwd),'exit':p.returncode,'seconds':time.monotonic()-start,'stopReason':reason}
 (raw/(name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(name,json.dumps(result),flush=True)
 if p.returncode:raise SystemExit(p.returncode if p.returncode>0 else 1)
 return result
