import json,subprocess,time
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');begin=time.monotonic()
while True:
 try:d=json.loads((s/'COUNT.json').read_text())
 except (FileNotFoundError,json.JSONDecodeError):d=None
 if d:
  assert d['differences']==0,'STOP: real-call counterexample'
  assert not any(p['failed'] for p in d['pieces']),'STOP: accepted count shard failure'
  print('accepted count progress',len(d['completedUnits']),d['calls'],d['differences'],flush=True)
  if d['complete']:break
 else:print('waiting for first complete count piece',round(time.monotonic()-begin),flush=True)
 assert time.monotonic()-begin<2400,'count wait exceeded bounded 40-minute budget'
 time.sleep(30)
with (s/'finalize-measurement.log').open('w') as log:r=subprocess.run(['python3',str(s/'finalize-measurement.py')],cwd='/workspace/adamic',stdout=log,stderr=subprocess.STDOUT)
print('measurement/apply pipeline exit',r.returncode,flush=True);raise SystemExit(r.returncode)
