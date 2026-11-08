from pathlib import Path
import os,subprocess,time,signal,json,sys,resource
s=Path('/workspace/scratch/core-finish');raw=s/'raw'
def run(name,args,cwd='/workspace/adamic',env=None):
 start=time.monotonic();last=start;size=-1;reason=None
 with (raw/(name+'.log')).open('w') as log:
  p=subprocess.Popen([str(x) for x in args],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   while p.poll() is None:
    current=(raw/(name+'.log')).stat().st_size
    if current!=size:size=current;last=time.monotonic()
    if time.monotonic()-last>1800 or time.monotonic()-start>6600:
     reason='30 minutes without output' if time.monotonic()-last>1800 else '110 minute bound';os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
     break
    time.sleep(1)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 result={'command':[str(x) for x in args],'cwd':str(cwd),'exit':p.returncode,'seconds':time.monotonic()-start,'stopReason':reason,'childrenPeakRssKiB':resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss}
 (raw/(name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(name,json.dumps(result),flush=True)
 if p.returncode:raise SystemExit(p.returncode if p.returncode>0 else 1)
 return result
if __name__=='__main__':run(sys.argv[1],sys.argv[3:],cwd=sys.argv[2])
