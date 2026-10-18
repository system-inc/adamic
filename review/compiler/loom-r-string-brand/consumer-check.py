import json,os,re,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=json.loads((out/'checks.json').read_text())
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
standard=['-count=1','-timeout','90s','-parallel','4','-json']
run('command-runtime-owner',['go','test','./internal/oracle','-run','^TestProduct_NativeRuntimesSanitized0$',*standard])
run('command-runtime-product',['go','test','./cmd/adamic','-run','^TestProduct_CommandRuntimeSanitized$',*standard])
run('command-consumers',['go','test','./cmd/adamic','-run','^TestNonNullExplainChecks$',*standard])
run('compiler-inputs',['go','test','./internal/compilerinputs','-run','^TestLoweringIs',*standard])
run('compiler-census',['go','test','./internal/compilercensus','-run','^TestDependencies',*standard])
run('meter-consumers',['go','test','./cmd/adamic-meter','-run','^TestOptionalAdaptationLeavesUnsafeContractsRejected$',*standard])
run('test262-consumers',['go','test','./cmd/adamic-test262','-run','^TestCompilerWorkerMatchesSubprocess$',*standard])
