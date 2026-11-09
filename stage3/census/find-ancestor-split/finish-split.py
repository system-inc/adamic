import json,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
while True:
 ready=True
 for side in ['a','b']:
  p=out/(side+'-entry-run.status.json')
  if p.exists():
   row=json.loads(p.read_text())
   if row['exit']!=0 or row['stopped'] is not None:sys.exit('run failed or stopped; no final commit')
  else:ready=False
  p=out/('checkpoint-'+side+'-entry-run.json')
  if p.exists():
   if json.loads(p.read_text())['exit']!=0:sys.exit('checkpoint failed; no final commit')
  else:ready=False
 if ready:break
 time.sleep(2)
for script in ['check-sources.py','render.py','seal-split.py']:
 rc=subprocess.run([sys.executable,str(out/script)],cwd=out.parents[2]).returncode
 print(script+' child_exit='+str(rc),flush=True)
 if rc:sys.exit(rc)
