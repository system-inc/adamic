import json,os,signal,subprocess,sys,time
from pathlib import Path
out=Path('/workspace/scratch/995spev')
name=sys.argv[1];args=sys.argv[2:];start=time.monotonic()
with (out/(name+'.log')).open('wb') as log:
 nested=os.environ.get('CENSUS_CHECKPOINT_CHILD')=='1'
 p=subprocess.Popen(args,stdout=log,stderr=log,start_new_session=not nested)
 reason=None
 try: rc=p.wait(timeout=1780)
 except subprocess.TimeoutExpired:
  reason='1780 second bound'
  if nested: p.terminate()
  else: os.killpg(p.pid,signal.SIGTERM)
  try: rc=p.wait(timeout=5)
  except subprocess.TimeoutExpired:
   if nested: p.kill()
   else: os.killpg(p.pid,signal.SIGKILL)
   rc=p.wait()
if not nested:
 try: os.killpg(p.pid,signal.SIGKILL)
 except ProcessLookupError: pass
status=dict(command=args,exit_code=rc,seconds=time.monotonic()-start,stopping_reason=reason,process_group_cleaned=not nested)
(out/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
print(name,'child_exit='+str(rc),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
