import json,os,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=[]
def run(label,command,expected=0,timeout=840):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:
  r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=timeout)
 row=dict(label=label,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3));results.append(row)
 (out/'checks.json').write_text(json.dumps(results,indent=2));print(label,r.returncode,row['seconds'],flush=True)
 assert r.returncode==expected,(label,r.returncode)
 events=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and 'Test' in e and e.get('Action') in ('pass','fail'):events.append(e)
 oversized=[{'test':e['Test'],'seconds':e.get('Elapsed',0)} for e in events if e.get('Elapsed',0)>=60 and not any(x['Test'].startswith(e['Test']+'/') for x in events)]
 if oversized:
  (out/(label+'-over-limit.json')).write_text(json.dumps(oversized,indent=2));raise AssertionError((label,oversized))
standard=['-count=1','-timeout','12m','-parallel','4','-p','1','-json']
run('stage3-update',['go','test','./stage3/fixtures','-run','^TestFixtures(Assertions|Namespaces)$',*standard,'-args','-update'])
run('stage3-all',['go','test','./stage3/fixtures/...',*standard])
run('counts-regenerate',['go','test','./internal/oracle','-run','TestCountsAreRecorded',*standard,'-args','-update-counts'])
run('lower-all',['go','test','./internal/lower/...',*standard])
run('native-all',['go','test','./internal/native/...',*standard])
run('oracle-all',['go','test','./internal/oracle/...',*standard])
run('readers-maps',['go','test','./internal/ir','-run','^Test(CallTargetReaders|EveryMapAProgramHoldsIsSorted)$','-count=1','-timeout','90s','-parallel','4','-json'])
run('parallel-guard',['go','test','./cmd/adamic-gate','-run','^TestEveryTestIsParallelOrSaysWhy$','-count=1','-timeout','90s','-parallel','4','-json'])
run('vet',['go','vet','./internal/lower/...','./internal/native/...','./internal/oracle/...','./stage3/fixtures/...'])
