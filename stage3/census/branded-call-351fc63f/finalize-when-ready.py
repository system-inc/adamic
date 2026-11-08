"""Bounded finalization after both paired runners, preserving a reviewable cutoff."""
import datetime,json,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
cutoff=datetime.datetime(2026,11,8,7,46,tzinfo=datetime.timezone.utc)
while True:
 complete=all((out/(a+'-complete.json')).exists() for a in ['A','B'])
 if complete:break
 if datetime.datetime.now(datetime.timezone.utc)>=cutoff:break
 time.sleep(2)
cmd=[sys.executable,str(out/'finalize.py')]
if not complete:cmd+=['--stop','--reason','The approximately two-hour environment budget was exhausted before the remaining entry head scan completed. Full-core and focused evidence is complete; no verdict is claimed for an incomplete entry pairing.']
rc=subprocess.run(cmd).returncode
print('finalizer child_exit='+str(rc),flush=True)
raise SystemExit(rc)
