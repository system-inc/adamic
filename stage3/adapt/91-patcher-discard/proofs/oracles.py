#!/usr/bin/env python3
"""Run control, validate it, then run adaptation 91, with no concurrent work."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

repo=Path(__file__).resolve().parents[4]; proof=Path(__file__).resolve().parent
sys.dont_write_bytecode=True
sys.path.insert(0,str(repo/'stage3/lane'))
from check import api_check, failed_tests
expected=json.loads((repo/'stage3/lane/expected.json').read_text())
sanctioned=json.loads((repo/'stage3/lane/sanctioned-api.json').read_text())
counts=dict(passing=106366,failing=1,pending=0)
records=[]
for label,tree in [('control',Path('/workspace/scratch/adapt91-control')),('after',Path('/workspace/scratch/adapt91-after'))]:
 out=Path('/workspace/scratch')/('adapt91-oracle-'+label)
 memory=Path('/sys/fs/cgroup/memory.events').read_text()
 with (proof/'logs'/('free-'+label+'.log')).open('w') as log:
  subprocess.run(['free','-m'],stdout=log,stderr=subprocess.STDOUT,check=True,timeout=10)
  log.write('\ncgroup memory.events before:\n'+memory);log.flush()
 command=['bash','stage3/oracle/run.sh',str(tree),str(out),'--workers','4','--limit-seconds','1800']
 start=time.monotonic()
 with (proof/'logs'/('oracle-'+label+'.log')).open('w') as log:
  result=subprocess.run(command,cwd=repo,stdout=log,stderr=subprocess.STDOUT,timeout=2400)
 report=json.loads((out/'report.json').read_text())
 current=Path('/sys/fs/cgroup/memory.events').read_text()
 row=dict(label=label,command=command,exit=result.returncode,seconds=round(time.monotonic()-start,3),counts=report.get('counts'),memoryBefore=memory,memoryAfter=current)
 records.append(row)
 (proof/'oracle-progress.json').write_text(json.dumps(records,indent=2)+'\n')
 assert current==memory,('memory event changed; stop',row)
 assert result.returncode==1 and report['counts']==counts,('oracle condition failed; stop',row)
 assert {k:v['exit'] for k,v in report['phases'].items()}==dict(install=0,build=0,tests=1)
 assert report['runners']=='all' and report['tests'] is None and report['workers']==4
 assert report['baseline_diffs']==expected['baseline_diffs']
 assert failed_tests((out/'tests.log').read_text())==expected['failed_tests']
 assert report['cause_counts']=={'timeout':0,'baseline-content':1,'baseline-missing':0,'exception':0,'other':0}
 lane=Path('/workspace/scratch')/('adapt91-api-'+label);lane.mkdir()
 (lane/'adapted-tree').symlink_to(tree,target_is_directory=True)
 (lane/'oracle').symlink_to(out,target_is_directory=True)
 errors,api=api_check(lane,sanctioned);assert not errors,errors
 row['sanctionedApi']=api
 (proof/('oracle-'+label+'.json')).write_text(json.dumps(report,indent=2)+'\n')
old=Path('/workspace/scratch/adapt91-oracle-control/baseline.diff').read_bytes()
new=Path('/workspace/scratch/adapt91-oracle-after/baseline.diff').read_bytes()
assert old==new,'baseline diff bytes moved'
(proof/'baseline.diff').write_bytes(new)
a=Path('/workspace/scratch/adapt91-control/built/local/typescript.d.ts').read_bytes()
b=Path('/workspace/scratch/adapt91-after/built/local/typescript.d.ts').read_bytes()
assert a==b
# Exercise the actual baseline-byte checker with a planted extra diff line.
lane=Path('/workspace/scratch/adapt91-api-after');diff=lane/'oracle/baseline.diff'
try:
 diff.write_bytes(new+b'\nextra baseline mutant\n')
 errors,_=api_check(lane,sanctioned)
 assert any('baseline.diff does not exactly describe' in e for e in errors)
finally: diff.write_bytes(new)
result=dict(status='pass',runs=records,baselineDiff=dict(byteIdentical=True,bytes=len(new),sha256=hashlib.sha256(new).hexdigest(),extraLineMutantCaught=True),api=dict(byteIdentical=True,bytes=len(a),sha256=hashlib.sha256(a).hexdigest()))
(proof/'oracle-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
