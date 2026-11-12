"""Bound each child to 29 minutes; report its own exit code and reap its group."""
import json,os,signal,subprocess,sys,time
from pathlib import Path
label=sys.argv[1];command=sys.argv[2:];start=time.monotonic()
p=subprocess.Popen(command,start_new_session=True)
stopped=None
try:
 try:rc=p.wait(timeout=1740)
 except subprocess.TimeoutExpired:
  stopped="29 minute child bound";os.killpg(p.pid,signal.SIGTERM)
  try:rc=p.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
finally:
 try:os.killpg(p.pid,signal.SIGKILL)
 except ProcessLookupError:pass
result={"command":command,"child_exit_code":rc,"seconds":time.monotonic()-start,"stopping_reason":stopped}
Path(label+"-status.json").write_text(json.dumps(result,indent=2)+"\n")
print(json.dumps(result),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
