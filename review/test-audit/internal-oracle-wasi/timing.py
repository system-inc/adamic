import subprocess,time,json,os
from pathlib import Path
root=Path('review/test-audit/internal-oracle-wasi')
rows=['TestWASIAgreesWithNode','TestWASIOracleCatchesMutants','TestWASIRunnerCatchesMutants','TestWASIEmission','TestWeakReadsUndefinedOnceFreed']
results=[]
for row in rows:
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$']
  start=time.monotonic()
  with open(root/f'time-{row}-{i}.log','w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
  elapsed=time.monotonic()-start
  events=[]
  for l in open(root/f'time-{row}-{i}.log'):
   try:events.append(json.loads(l))
   except:pass
  end=[e for e in events if e.get('Action') in ['pass','fail'] and 'Test' not in e]
  results.append(dict(test=row,run=i,command=cmd,status=r.returncode,wall=elapsed,binary_seconds=end[-1].get('Elapsed') if end else None))
  (root/'timings.json').write_text(json.dumps(results,indent=2))
  # A timeout is cooked; do not spend two more runs on the same over-budget row.
  if any('test timed out' in e.get('Output','') for e in events):break
  if r.returncode:raise SystemExit('red baseline '+row)
