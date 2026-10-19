import json, subprocess, time, re
from pathlib import Path
out=Path(__file__).resolve().parent
focus=json.loads((out/'focus.json').read_text())
results=json.loads((out/'checks.json').read_text())
def run(label,command,expected=0):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:
  p=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=240)
 result=dict(label=label,command=command,exit=p.returncode,seconds=round(time.monotonic()-start,3))
 results.append(result);(out/'checks.json').write_text(json.dumps(results,indent=2))
 print(label,p.returncode,result['seconds'],flush=True)
 if p.returncode!=expected:raise SystemExit(1)
paths='internal/oracle/testdata/('+'|'.join(re.escape(p.split("/")[3]) for p in focus['fixtures'])+')$'
run('fixtures-rows',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/'+paths,'-count=1','-timeout','90s','-parallel','4','-json'])
# Measure only the change's fifteen fixtures, then write the table once, preserving every unmeasured main row.
run('counts-regeneration',['go','test','./internal/oracle','-run','^TestCountsAreRecorded_[0-9]+$/'+paths,'-count=1','-timeout','90s','-parallel','4','-json','-args','-update-counts'])
run('counts-table',['go','test','./internal/oracle','-run','^TestCountsAreRecordedTable$','-count=1','-timeout','90s','-json'])
for case in json.loads((out/'mutants/cases.json').read_text()):
 label='mutant-'+case['topic']
 run(label,['go','test','-overlay',case['overlay'],*case['packages'],'-run',case['pattern'],'-count=1','-timeout','90s','-parallel','4','-json'],1)
 log=(out/(label+'.log')).read_text()
 if '"Action":"fail","Package"' not in log or '"Test":' not in log or '[build failed]' in log or 'no tests to run' in log:
  raise SystemExit('invalid mutant failure '+label)
