import subprocess,time,json,os
from pathlib import Path
p=Path(__file__).resolve().parent
results=[]
env=os.environ.copy();env['ADAMIC_ESTREE_LIBRARY']='/tmp/u088/library'
for i,(name,regex) in enumerate(json.loads((p/'groups.json').read_text())):
 for repeat in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex]
  start=time.monotonic();log=p/f'timing-{i:02d}-{repeat}.log'
  with log.open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  events=[]
  for line in log.read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  result={'test':name,'repeat':repeat,'command':cmd,'wall':time.monotonic()-start,'exit':r.returncode,'binary_seconds':next((e.get('Elapsed') for e in reversed(events) if not e.get('Test') and e.get('Action') in ['pass','fail']),None),'failures':[e.get('Test') for e in events if e.get('Test') and e.get('Action')=='fail']}
  results.append(result);(p/'timings.json').write_text(json.dumps(results,indent=2))
  print(name,repeat,result['exit'],result['binary_seconds'],flush=True)
  if r.returncode:raise SystemExit('baseline row failed; stop before mutants')
