import json, os, signal, subprocess, sys, time
from pathlib import Path
name=sys.argv[1]; command=sys.argv[2:]; out=Path(__file__).resolve().parent
started=time.monotonic()
with (out/(name+'.log')).open('wb') as log:
 p=subprocess.Popen(command,stdout=log,stderr=log,start_new_session=True)
 try: rc=p.wait(timeout=1700)
 except subprocess.TimeoutExpired:
  os.killpg(p.pid,signal.SIGTERM)
  try:rc=p.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
result={'command':command,'child_exit_code':rc,'seconds':time.monotonic()-started}
(out/(name+'.json')).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
