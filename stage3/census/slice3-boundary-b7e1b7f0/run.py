import json,os,signal,subprocess,sys,time
from pathlib import Path
prefix=Path(sys.argv[1]);limit=float(sys.argv[2]);assert 0<limit<=1500
command=sys.argv[3:];prefix.parent.mkdir(parents=True,exist_ok=True);start=time.monotonic();timed=False
with prefix.with_suffix('.log').open('wb') as log:
 child=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 def stop(sig,frame):
  try:os.killpg(child.pid,signal.SIGTERM)
  except ProcessLookupError:pass
 signal.signal(signal.SIGTERM,stop);signal.signal(signal.SIGINT,stop)
 while True:
  pid,status,usage=os.wait4(child.pid,os.WNOHANG)
  if pid:break
  if time.monotonic()-start>limit:
   timed=True;os.killpg(child.pid,signal.SIGKILL);pid,status,usage=os.wait4(child.pid,0);break
  time.sleep(.2)
 code=os.waitstatus_to_exitcode(status);child.returncode=code
 try:os.killpg(child.pid,signal.SIGKILL)
 except ProcessLookupError:pass
result={'command':command,'cwd':os.getcwd(),'child_exit':code,'timed_out':timed,'limit_seconds':limit,'elapsed_seconds':time.monotonic()-start,'peak_rss_kib':usage.ru_maxrss,'user_seconds':usage.ru_utime,'system_seconds':usage.ru_stime,'environment':{k:v for k,v in os.environ.items() if k.startswith(('LATENT_','BOUNDARY_')) or k in ('GOMAXPROCS','GOMEMLIMIT','ADAMIC_TYPESCRIPT_SOURCE')}}
prefix.with_suffix('.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True);sys.exit(code if code>=0 else 128-code)
