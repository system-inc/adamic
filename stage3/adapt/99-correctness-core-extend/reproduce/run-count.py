import json,time,subprocess,os
from pathlib import Path
s=Path('/workspace/scratch/core-extend')
while not (s/'count-preparation.json').exists() or 'harness' not in json.loads((s/'count-preparation.json').read_text()):
 print('waiting for complete recut counting harness',flush=True);time.sleep(20)
assert all(p['exit']==0 for p in json.loads((s/'count-preparation.json').read_text()).values())
with (s/'verify-count-inputs.log').open('w') as log:subprocess.run(['python3',str(s/'verify-count-inputs.py')],stdout=log,stderr=subprocess.STDOUT,check=True)
with (s/'count-pieces.log').open('w') as log:r=subprocess.run(['python3',str(s/'count-pieces.py')],stdout=log,stderr=subprocess.STDOUT)
print('recut count exit',r.returncode,flush=True);raise SystemExit(r.returncode)
