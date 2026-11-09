import subprocess,time,sys,json
from pathlib import Path
out=Path(__file__).resolve().parent
while not (out/'apply.status.json').exists():time.sleep(2)
status=json.loads((out/'apply.status.json').read_text())
if status['exit']!=0:sys.exit('apply failed; census not started')
rc=subprocess.run([sys.executable,str(out/'check-sources.py')]).returncode
print('source check wrapper_exit='+str(rc),flush=True)
if rc:sys.exit(rc)
rc=subprocess.run([sys.executable,str(out/'checkpoint.py'),'adapted']).returncode
print('adapted checkpoint wrapper_exit='+str(rc),flush=True)
if rc:sys.exit(rc)
while not (out/'provenance.json').exists():
 for p in out.glob('*-build.status.json'):
  if json.loads(p.read_text())['exit']!=0:sys.exit('build failed; census not started')
 time.sleep(2)
rc=subprocess.run([sys.executable,str(out/'census-split.py')]).returncode
print('census controller wrapper_exit='+str(rc),flush=True)
sys.exit(rc)
