"""Retain a bounded command's exact child exit, output, arguments and timing."""
import json,os,pathlib,signal,subprocess,sys,time
name=sys.argv[1];command=sys.argv[2:];root=pathlib.Path(__file__).resolve().parent
started=time.monotonic()
with (root/(name+'.stdout')).open('w') as out, (root/(name+'.stderr')).open('w') as err:
 child=subprocess.Popen(command,stdout=out,stderr=err,start_new_session=True)
 try:code=child.wait(timeout=1500)
 except subprocess.TimeoutExpired:
  os.killpg(child.pid,signal.SIGTERM)
  try:code=child.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);code=child.wait()
record=dict(command=command,cwd=os.getcwd(),child_exit=code,seconds=time.monotonic()-started)
(root/(name+'.status.json')).write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record),flush=True)
sys.exit(code if code>=0 else 128-code)
