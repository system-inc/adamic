import subprocess,pathlib,json,time,os,signal
S=pathlib.Path('/workspace/scratch/sys-timer-stack');ROOT=pathlib.Path('/workspace/adamic');env={**os.environ,'NODE_PATH':str(ROOT/'stage3/api/node_modules'),'ADAMIC_TYPESCRIPT_SOURCE':'/workspace/typescript-source'}
def run(name,args,cwd=ROOT,extra=None,limit=1800):
 started=time.monotonic();last=started;old=None;log=S/(name+'.log');stopped=None
 with log.open('w') as stream:
  p=subprocess.Popen(list(map(str,args)),cwd=cwd,env={**env,**(extra or {})},stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
  while p.poll() is None:
   signature=(log.stat().st_size,log.stat().st_mtime_ns)
   now=time.monotonic()
   if signature!=old:last=now;old=signature
   if now-last>=1800 or now-started>=limit:
    stopped='30 minutes without output' if now-last>=1800 else 'bounded piece elapsed limit';os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
    break
   time.sleep(.5)
  code=p.wait()
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 result={'name':name,'command':list(map(str,args)),'cwd':str(cwd),'exit':code,'seconds':time.monotonic()-started,'stoppingReason':stopped}
 (S/(name+'-status.json')).write_text(json.dumps(result,indent=2)+'\n');print(name+' child_exit='+str(code),flush=True);return code
