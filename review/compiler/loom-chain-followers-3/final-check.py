import json,os,re,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
focus=json.loads((out/'focus.json').read_text());results=json.loads((out/'checks.json').read_text())
def run(label,command,expected=0,minimum=1):
 start=time.monotonic();log=out/(label+'.log')
 with log.open('w') as stream:
  completed=subprocess.run(command,stdout=stream,stderr=subprocess.STDOUT,timeout=240)
 result=dict(label=label,command=command,exit=completed.returncode,seconds=round(time.monotonic()-start,3))
 results.append(result);(out/'checks.json').write_text(json.dumps(results,indent=2))
 print(label,result['exit'],result['seconds'],flush=True)
 text=log.read_text();events=[]
 for line in text.splitlines():
  if line.startswith('{'):
   try:events.append(json.loads(line))
   except json.JSONDecodeError:pass
 tests=[r for r in events if r.get('Action') in ['pass','fail'] and 'Test' in r]
 assert completed.returncode==expected,(label,completed.returncode)
 assert len(tests)>=minimum,(label,'matched no tests')
 assert not any(r.get('Elapsed',0)>=60 for r in tests),(label,'test crossed 60 seconds')
 if expected==1:
  assert '[build failed]' not in text and any(r['Action']=='fail' for r in tests),(label,'not a test catcher')
 return tests
fixture_paths=focus['fixtures']
# Each slash is a Go test-name component; only the final component has alternatives.
paths='internal/oracle/testdata/('+'|'.join(sorted({re.escape(p.split('/')[3]) for p in fixture_paths}))+')$'
standard=['-count=1','-timeout','90s','-parallel','4','-json']
run('final-parallel-rows',['go','test','./internal/lower','-run','^Test(GenericFunctionValueIdentityCallsAreNotYet|PhantomRefusals)$',*standard])
run('final-fixtures',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/'+paths,*standard],minimum=15)
run('final-readers',['go','test','./internal/ir','-run','^TestCallTargetReaders$',*standard])
run('final-parallel-guard',['go','test','./cmd/adamic-gate','-run','^TestEveryTestIsParallelOrSaysWhy$',*standard])
run('final-vet',['go','vet','./internal/flow','./internal/ir','./internal/javascript','./internal/lower','./internal/native','./internal/oracle','./internal/fresh'],minimum=0)
