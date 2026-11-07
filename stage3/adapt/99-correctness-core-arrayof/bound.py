from pathlib import Path
import json,os,signal,subprocess,sys,time
s=Path('/workspace/scratch/arrayof');label=sys.argv[1];limit=float(sys.argv[2]);command=sys.argv[3:];started=time.monotonic();timedout=False;killed=[]
with (s/(label+'.log')).open('w') as log:
 p=subprocess.Popen(command,cwd='/workspace/adamic',stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=p.wait(timeout=limit)
 except subprocess.TimeoutExpired:
  timedout=True;rows=[list(map(int,l.split())) for l in subprocess.check_output(['ps','-eo','pid=,ppid='],text=True).splitlines()];ids={p.pid}
  while True:
   expanded=ids|{pid for pid,parent in rows if parent in ids}
   if expanded==ids:break
   ids=expanded
  for pid in sorted(ids,reverse=True):
   try:os.killpg(pid,signal.SIGKILL)
   except ProcessLookupError:pass
   try:os.kill(pid,signal.SIGKILL)
   except ProcessLookupError:pass
  killed=sorted(ids);p.wait();code=124
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
result=dict(exit=code,seconds=time.monotonic()-started,limitSeconds=limit,timedOut=timedout,command=command,processGroupCleaned=True,killedOnTimeout=killed);(s/(label+'-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True);sys.exit(code)
