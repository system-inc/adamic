import json,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=json.loads((out/'checks.json').read_text())
def run(label,command,timeout=840):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=timeout)
 row=dict(label=label,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3));results.append(row);(out/'checks.json').write_text(json.dumps(results,indent=2));print(label,r.returncode,row['seconds'],flush=True)
 events=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and 'Test' in e and e.get('Action') in ('pass','fail'):events.append(e)
 large=[e for e in events if e.get('Elapsed',0)>=60 and not any(x.get('Package')==e.get('Package') and x['Test'].startswith(e['Test']+'/') for x in events)]
 assert not large,(label,large)
 assert r.returncode==0,(label,r.returncode)
standard=['-count=1','-timeout','12m','-parallel','4','-p','1','-json']
for name in ['ir','javascript','flow','fresh']:
 run(name+'-all-ready',['go','test','./internal/'+name+'/...',*standard])
run('stage3-fixtures',['go','test','./stage3/fixtures/...',*standard])
run('readers-maps',['go','test','./internal/ir','-run','^Test(CallTargetReaders.*|EveryMapAProgramHoldsIsSorted)$','-count=1','-timeout','90s','-parallel','4','-json'],180)
run('parallel-guard',['go','test','./cmd/adamic-gate','-run','^TestEveryTestIsParallelOrSaysWhy$','-count=1','-timeout','90s','-parallel','4','-json'],180)
run('vet',['go','vet','./internal/ir/...','./internal/lower/...','./internal/native/...','./internal/oracle/...','./internal/javascript/...','./internal/flow/...','./internal/fresh/...','./stage3/fixtures/...'],240)
