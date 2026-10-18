import json,os,re,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
focus=json.loads((out/'focus.json').read_text());results=[]
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
fixture_paths=focus['fixtures']+focus['main_witnesses']
# Each slash is a Go test-name component; only the final component has alternatives.
paths='internal/oracle/testdata/('+'|'.join(sorted({re.escape(p.split('/')[3]) for p in fixture_paths}))+')$'
standard=['-count=1','-timeout','90s','-parallel','4','-json']
run('counts-regeneration',['go','test','./internal/oracle','-run','^TestCountsAreRecorded_[0-9]+$/'+paths,*standard,'-args','-update-counts'],minimum=22)
run('counts-table',['go','test','./internal/oracle','-run','^TestCountsAreRecordedTable$',*standard])
for package,names in focus['tests'].items():
 run(package.split('/')[-1]+'-focus',['go','test','./'+package,'-run','^('+'|'.join(names)+')$',*standard],minimum=len(names))
run('fixtures',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/'+paths,*standard],minimum=22)
run('stage3',['go','test','./stage3/fixtures','-run','^TestFixtures(Assertions|Cycles|Namespaces)$/(assertions|cycles|namespaces)/(16_unknown_chain.a|17_brand_upcast.a|04_directory_callback|05_safe_load_read|06_import_order_mutant|02_jsx_names.a|03_react_names.a)',*standard],minimum=7)
run('readers',['go','test','./internal/ir','-run','^TestCallTargetReaders$',*standard])
run('vet',['go','vet','./internal/flow','./internal/ir','./internal/javascript','./internal/lower','./internal/native','./internal/oracle'],minimum=0)
for case in json.loads((out/'mutants/cases.json').read_text()):
 run('mutant-'+case['topic'],['go','test','-overlay',case['overlay'],*case['packages'],'-run',case['pattern'],*standard],1)
print('All focused checks and 14 source mutants completed; 7 IR output mutants are held by oracle-focus.',flush=True)
