import json,os,signal,subprocess,time,sys
from pathlib import Path
scratch=Path('/workspace/scratch/vqzty8t')
def run(name,cmd,env=None,watch=(),limit=1490,cwd='/workspace/view-pairs'):
 started=time.monotonic();last=started;previous=None;stopped=None
 paths=[scratch/(name+'.log'),*[Path(p) for p in watch]]
 with paths[0].open('wb') as log:
  p=subprocess.Popen(cmd,cwd=cwd,stdout=log,stderr=log,env=env,start_new_session=True)
  try:
   while p.poll() is None:
    sig=[(f.stat().st_size,f.stat().st_mtime_ns) if f.exists() else None for f in paths]
    now=time.monotonic()
    if sig!=previous:previous=sig;last=now
    if now-last>=1490 or now-started>=limit:
     stopped='configured elapsed bound' if now-started>=limit else 'silence bound'
     os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=5)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
     break
    time.sleep(1)
   rc=p.wait()
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
 status=dict(command=cmd,exit_code=rc,stopping_reason=stopped,seconds=time.monotonic()-started,process_group_cleaned=True)
 (scratch/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
 print(name+' child_exit='+str(rc),flush=True)
 return status
if __name__=='__main__':
 r=run(sys.argv[1],sys.argv[2:]);sys.exit(r['exit_code'] if r['exit_code']>=0 else 128-r['exit_code'])
