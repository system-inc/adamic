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
run('fresh-consumers',['go','test','./internal/fresh','-run','^TestFreshCorpusRemainder$',*standard])
run('native-consumers',['go','test','./internal/native','-run','^Test(PassThroughsAreNotConsumers|InheritanceMemoryPlans)$',*standard])
run('javascript-consumers',['go','test','./internal/javascript','-run','^TestViewMixedUnionSelectsNumber$',*standard])
run('guards-ir',['go','test','./internal/ir','-run','^TestEveryMapAProgramHoldsIsSorted$',*standard])
run('guards-runtime',['go','test','./internal/oracle','-run','^TestRuntimeSetsAreEachUsed$',*standard])
run('guards-registration',['go','test','./internal/lower','-run','^TestNativeAgreementRowsNameRegisteredPrograms$',*standard])
run('guards-parallel',['go','test','./cmd/adamic-gate','-run','^TestEveryTestIsParallelOrSaysWhy$',*standard])
run('readers',['go','test','./internal/ir','-run','^TestCallTargetReaders$',*standard])
run('vet',['go','vet','./internal/flow','./internal/ir','./internal/javascript','./internal/lower','./internal/native','./internal/oracle','./internal/fresh'],minimum=0)
run('mutant-phantom-fresh',['go','test','-overlay',str(out/'mutants/phantom-fresh.json'),'./internal/fresh','-run','^TestFreshCorpusRemainder$',*standard],1)
for case in json.loads((out/'mutants/cases.json').read_text()):
 run('mutant-'+case['topic'],['go','test','-overlay',case['overlay'],*case['packages'],'-run',case['pattern'],*standard],1)
print('All focused checks and 14 source mutants completed; 7 IR output mutants are held by oracle-focus.',flush=True)
