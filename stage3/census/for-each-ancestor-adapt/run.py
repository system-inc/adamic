#!/usr/bin/env python3
"""Bound one child to 29 minutes; preserve its own exit status."""
import json, os, signal, subprocess, sys, time
out, *command = sys.argv[1:]
start=time.monotonic()
with open(out+'.log','wb') as log:
 p=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 timed=False
 try: code=p.wait(timeout=1740)
 except subprocess.TimeoutExpired:
  timed=True; os.killpg(p.pid,signal.SIGTERM)
  try: code=p.wait(timeout=5)
  except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL); code=p.wait()
result=dict(command=command,child_exit=code,timed_out=timed,seconds=time.monotonic()-start)
with open(out+'.status.json','w') as f: json.dump(result,f,indent=2); f.write('\n')
print(json.dumps(result),flush=True)
sys.exit(code if code>=0 else 128-code)
