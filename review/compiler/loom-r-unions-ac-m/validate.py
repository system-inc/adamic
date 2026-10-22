import json,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=json.loads((out/'checks.json').read_text()) if (out/'checks.json').exists() else []
def run(label,command,expected=0,timeout=840):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=timeout)
 row=dict(label=label,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3));results.append(row);(out/'checks.json').write_text(json.dumps(results,indent=2));print(label,r.returncode,row['seconds'],flush=True)
 events=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and 'Test' in e and e.get('Action') in ('pass','fail'):events.append(e)
 large=[e for e in events if e.get('Elapsed',0)>=60 and not any(x['Test'].startswith(e['Test']+'/') for x in events)]
 assert not large,(label,large)
 assert r.returncode==expected,(label,r.returncode)
standard=['-count=1','-timeout','12m','-parallel','4','-p','1','-json']
# Completed setup/build/stage3/counts/meter steps are retained in checks.json.
run('stage1-touched-ready',['go','test','./stage1/cohere/selector/...','./stage1/cohere/values/...',*standard])
run('flow-fresh',['go','test','./internal/flow/...','./internal/fresh/...',*standard])
run('readers-maps',['go','test','./internal/ir','-run','^Test(CallTargetReaders|EveryMapAProgramHoldsIsSorted)$','-count=1','-timeout','90s','-parallel','4','-json'])
run('parallel-guard',['go','test','./cmd/adamic-gate','-run','^TestEveryTestIsParallelOrSaysWhy$','-count=1','-timeout','90s','-parallel','4','-json'])
run('vet',['go','vet','./cmd/adamic-meter/...','./internal/lower/...','./internal/native/...','./internal/oracle/...','./stage1/cohere/selector/...','./stage1/cohere/values/...','./internal/flow/...','./internal/fresh/...','./internal/ir/...','./stage3/fixtures/...'])
