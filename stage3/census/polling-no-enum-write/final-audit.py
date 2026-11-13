#!/usr/bin/env python3
"""Audit the retained raw acceptance artifacts and actual before/after files."""
from pathlib import Path
import hashlib,json,subprocess
here=Path(__file__).resolve().parent
repo=Path('/workspace/adamic');before=Path('/workspace/scratch/polling-before');after=Path('/workspace/scratch/polling-after')
checks=json.loads((here/'final-checks.json').read_text())
assert all(row['exit']==(1 if name=='restoring-mutant' else 0) for name,row in checks.items())
for label in ['before','after']:
 assert '"roots":79,"diagnostics":0' in (here/(label+'-stock-corrected.log')).read_text()
assert all(row['exit']==0 for row in json.loads((here/'build-exits.json').read_text()))
pin=json.loads((repo/'stage3/adapt/99-correctness-polling-no-enum-write/owners.json').read_text())[0]
b=(before/pin['file']).read_bytes();a=(after/pin['file']).read_bytes();ending=b'\r\n' if b'\r\n' in b else b'\n'
old=pin['before'].encode().replace(b'\n',ending);new=pin['after'].encode().replace(b'\n',ending)
assert b.count(old)==1 and b.replace(old,new)==a
scope=json.loads((here/'source-scope.json').read_text())
for label,tree in [('before',before),('after',after)]:
 for name,digest in scope[label].items(): assert hashlib.sha256((tree/name).read_bytes()).hexdigest()==digest,name
assert scope['changed']==[pin['file']]
for name in ['built/local/typescript.d.ts','tests/baselines/reference/api/typescript.d.ts']:
 assert (before/name).read_bytes()==(after/name).read_bytes(),name
census=json.loads((here/'after-census.json').read_text());mutant=json.loads((here/'mutant-census.json').read_text())
assert census['converged'] and census['counts']['writes']==0
assert census['counts']['classes']=={'a':36,'b':15,'c':0}
assert mutant['converged'] and mutant['counts']['writes']==1
assert mutant['writes'][0]['enum']=='PollingInterval'
assert mutant['writes'][0]['expression']=='levels[level] = customLevels![level] || levels[level]'
assert (here/'before-no-vars.log').read_bytes()==(here/'after-no-vars.log').read_bytes()
witness=json.loads((here/'witness-assertions.json').read_text());assert witness['before']==[17,41,89] and witness['after']==[250,500,2000]
rows=json.loads((here/'shard-exits.json').read_text());assert len(rows)==398
assert all(row['exit']==0 and row['status']=='pass' and row['seconds']<1800 for row in rows)
counts={k:sum(row['counts'][k] for row in rows) for k in ['pass','fail','skip']}
assert counts=={'pass':106366,'fail':1,'skip':0}
reports=[json.loads((here/'shard-logs'/(row['name']+'.json')).read_text()) for row in rows]
failures=[name for report in reports for name in report['failed_tests']]
paths=[name for report in reports for name in report['baseline_diffs']]
expected=json.loads((repo/'stage3/lane/expected.json').read_text())
assert failures==expected['failed_tests'] and paths==expected['baseline_diffs']
tracked=subprocess.check_output(['git','-C',str(before),'ls-files','-z']).decode().split('\0')
for name in filter(None,tracked):
 if name!=pin['file'] and (before/name).is_file(): assert (before/name).read_bytes()==(after/name).read_bytes(),name
result=dict(status='pass',shards=len(rows),counts=counts,max_shard_seconds=max(row['seconds'] for row in rows),failed_tests=failures,baseline_diffs=paths,source_files=len(scope['after']),tracked_files=json.loads((here/'tracked-file-scope.json').read_text())['files'],changed=[pin['file']])
(here/'final-audit.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
