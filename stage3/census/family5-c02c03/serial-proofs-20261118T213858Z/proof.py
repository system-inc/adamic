import os,sys,subprocess,json,time,datetime,signal
from pathlib import Path
out=Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True);cmd=sys.argv[2:];start=time.monotonic();utc=lambda:datetime.datetime.now(datetime.timezone.utc).isoformat();began=utc()
with (out/'stdout.log').open('wb') as stdout,(out/'stderr.log').open('wb') as stderr:
 p=subprocess.Popen(cmd,stdout=stdout,stderr=stderr,start_new_session=True)
 child={'pid':p.pid,'supervisor_pid':os.getpid(),'started_utc':began,'command':cmd,'proc_stat_at_launch':Path('/proc/'+str(p.pid)+'/stat').read_text()}
 (out/'child.json').write_text(json.dumps(child,indent=2)+'\n')
 while True:
  try:code=p.wait(timeout=50);break
  except subprocess.TimeoutExpired:print('RUNNING '+str(out)+' '+str(round(time.monotonic()-start,1))+'s',flush=True)
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:pass
r={'command':cmd,'cwd':os.getcwd(),'started_utc':began,'ended_utc':utc(),'environment':{k:v for k,v in os.environ.items() if k.startswith(('ADAMIC_','LATENT_','CENSUS_','FAMILY5_'))},'exit_code':code,'wall_seconds':round(time.monotonic()-start,3)}
(out/'status.json').write_text(json.dumps(r,indent=2)+'\n');print(('PASS ' if code==0 else 'FAIL ')+str(out)+' real_child_exit='+str(code),flush=True);sys.exit(code if code>=0 else 128-code)
