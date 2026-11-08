import os,sys,time,json,signal,subprocess,datetime,hashlib,socket
from pathlib import Path
out=Path(__file__).resolve().parent
raw=out/'raw'; raw.mkdir(exist_ok=True)
repo=out.parents[2]
def run(name,cmd,cwd=None,env=None,watch=(),limit=5400):
 statusfile=raw/(name+'-status.json')
 if statusfile.exists():
  old=json.loads(statusfile.read_text())
  if old['exit_code']==0 and old['command']==cmd:
   print(name+' reused successful child_exit=0',flush=True);return old
 started=time.time();last=started;prev=None;stop=None;usage=None
 with (raw/(name+'.log')).open('wb') as log:
  p=subprocess.Popen(cmd,cwd=cwd or repo,env=env,stdout=log,stderr=log,start_new_session=True)
  try:
   while True:
    pid,status,usage=os.wait4(p.pid,os.WNOHANG)
    if pid:p.returncode=os.waitstatus_to_exitcode(status);break
    now=time.time();paths=[raw/(name+'.log'),*map(Path,watch)]
    sig=[(f.stat().st_size,f.stat().st_mtime_ns) if f.exists() else None for f in paths]
    if sig!=prev:prev=sig;last=now
    if now-last>=1800 or now-started>=limit:
     stop='30 minutes without output' if now-last>=1800 else 'elapsed limit'
     os.killpg(p.pid,signal.SIGTERM)
     for _ in range(10):
      pid,status,usage=os.wait4(p.pid,os.WNOHANG)
      if pid:break
      time.sleep(1)
     if not pid:os.killpg(p.pid,signal.SIGKILL);pid,status,usage=os.wait4(p.pid,0)
     p.returncode=os.waitstatus_to_exitcode(status);break
    time.sleep(1)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
 status=dict(pid=p.pid,command=cmd,cwd=str(cwd or repo),started_utc=datetime.datetime.fromtimestamp(started,datetime.timezone.utc).isoformat(),ended_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),seconds=time.time()-started,exit_code=p.returncode,peak_rss_kib=usage.ru_maxrss,stopping_reason=stop,process_group_cleaned=True)
 (raw/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n');print(name+' child_exit='+str(p.returncode),flush=True)
 if p.returncode:raise RuntimeError(name+' failed: '+str(p.returncode))
 return status
if __name__=='__main__':run(sys.argv[1],sys.argv[2:])
