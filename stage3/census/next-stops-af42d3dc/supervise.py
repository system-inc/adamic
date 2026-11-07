"""Bound subprocess silence, retain its own exit code, and reap its process group."""
import json,os,signal,subprocess,time
from pathlib import Path
scratch=Path('/workspace/scratch/after-wave')
def run(name,cmd,env=None,watch=(),limit=1800,cwd='/workspace/adamic'):
 started=time.monotonic(); last=started; previous=None; stopped=None
 paths=[scratch/(name+'.log'),*[Path(p) for p in watch]]
 with paths[0].open('wb') as log:
  p=subprocess.Popen(cmd,cwd=cwd,stdout=log,stderr=log,env=env,start_new_session=True)
  (scratch/(name+'-pid.json')).write_text(json.dumps({'pid':p.pid,'command':cmd,'started_epoch':time.time()})+'\n')
  try:
   while p.poll() is None:
    signature=[(f.stat().st_size,f.stat().st_mtime_ns) if f.exists() else None for f in paths]
    now=time.monotonic()
    if signature!=previous: previous=signature;last=now
    if now-last>=1800 or now-started>=limit:
     stopped='30 minutes without output' if now-last>=1800 else 'configured elapsed-time bound'
     os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
     break
    time.sleep(1)
   rc=p.wait()
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
 status={'command':cmd,'exit_code':rc,'stopping_reason':stopped,'seconds':time.monotonic()-started,'process_group_cleaned':True}
 (scratch/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
 print(name+' child_exit='+str(rc),flush=True)
 return status
if __name__=='__main__':
 import sys
 result=run(sys.argv[1],sys.argv[2:],limit=7200)
 sys.exit(0 if result['exit_code']==0 else 1)
