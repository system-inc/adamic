import subprocess,os,signal,time,json
from pathlib import Path
S=Path('/workspace/scratch/elementat-declaration')
def run(name,args,cwd='/workspace/adamic',env=None,limit=1400):
 start=time.monotonic()
 with (S/(name+'.log')).open('wb') as log:
  p=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  timed=False
  try:p.wait(timeout=limit)
  except subprocess.TimeoutExpired:timed=True
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 row=dict(name=name,args=args,exit=p.returncode,timedOut=timed,seconds=time.monotonic()-start)
 with (S/'commands.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(json.dumps(row),flush=True)
 return row
if __name__=='__main__':
 import sys
 r=run(sys.argv[1],sys.argv[2:]);raise SystemExit(r['exit'] if r['exit']>=0 else 1)
