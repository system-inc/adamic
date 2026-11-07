import json,os,pathlib,signal,subprocess,sys,time
label=sys.argv[1]; command=sys.argv[2:]; root=pathlib.Path('/workspace/scratch/or-predicate'); started=time.time()
log=root/(label+'.log'); result=root/(label+'-result.json')
with log.open('wb') as out:
 p=subprocess.Popen(command,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
 last_size=-1; last_output=started; stopped=False
 while p.poll() is None:
  size=log.stat().st_size
  if size!=last_size: last_output=time.time();last_size=size
  if time.time()-last_output>1800:
   stopped=True;os.killpg(p.pid,signal.SIGTERM)
   break
  try:p.wait(timeout=5)
  except subprocess.TimeoutExpired:pass
 if stopped:
  try:p.wait(timeout=10)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
 code=p.returncode
 try:os.killpg(p.pid,signal.SIGKILL)
 except ProcessLookupError:pass
result.write_text(json.dumps(dict(command=command,exit_code=code,seconds=time.time()-started,stopped_after_30_minutes_without_output=stopped),indent=2)+'\n')
print(label+' child exit '+str(code),flush=True)
sys.exit(code if code>=0 else 128-code)
