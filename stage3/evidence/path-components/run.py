import json,os,signal,subprocess,sys,time
from pathlib import Path
out=Path(__file__).parent
name=sys.argv[1];cmd=sys.argv[2:];started=time.monotonic()
with (out/(name+'.log')).open('wb') as log:
 p=subprocess.Popen(cmd,stdout=log,stderr=log,start_new_session=True)
 stopped=None
 try: rc=p.wait(timeout=1740)
 except subprocess.TimeoutExpired:
  stopped='1740 second limit';os.killpg(p.pid,signal.SIGTERM)
  try:rc=p.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
 result=dict(command=cmd,exit_code=rc,stopping_reason=stopped,seconds=time.monotonic()-started)
 (out/(name+'-status.json')).write_text(json.dumps(result,indent=2)+'\n')
 print(name+' child_exit='+str(rc),flush=True)
 sys.exit(rc if rc>=0 else 128-rc)
