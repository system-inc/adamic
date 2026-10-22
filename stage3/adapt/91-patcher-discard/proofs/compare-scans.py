#!/usr/bin/env python3
import json
from pathlib import Path
import sys
import re
from extract import extract
before, after, mutant, output = map(Path, sys.argv[1:])
def load(file):
 records=[json.loads(line) for line in file.read_text().splitlines()]
 assert records[0]['latent_mode']=='void-widening-only-source-scan'
 assert len(records[1:])==79
 assert all(f['kind']!='panic' for r in records for f in r.get('findings') or [])
 return extract(file)
a,b,m=map(load,(before,after,mutant))
def all_findings(file):
 result=set()
 for line in file.read_text().splitlines():
  for f in json.loads(line).get('findings') or []:
   result.add(tuple(re.sub(r'/workspace/scratch/adapt91-[^/]+/', '<tree>/', str(f.get(k,''))) for k in ('kind','where','reason','text')))
 return result
raw_a,raw_b,raw_m=map(all_findings,(before,after,mutant))
assert not raw_b-raw_a, 'new source-scan finding of any kind'
assert not raw_b-raw_m, 'mutant lost an existing finding'

def key(r): return (r['file'],r['line'],r['column'])
def delta(old,new):
 x={key(r):r for r in old}; y={key(r):r for r in new}
 assert all(x[k]==y[k] for k in x.keys()&y.keys()), 'existing row changed'
 return [x[k] for k in x.keys()-y.keys()],[y[k] for k in y.keys()-x.keys()]
removed,added=delta(a,b)
assert not added
assert len(removed)==4
assert {(r['file'],r['line']) for r in removed}=={('src/compiler/factory/nodeFactory.ts',1164),('src/compiler/utilities.ts',8582)}
assert all(not any(key(x)==key(r) for x in b) for r in removed)
assert len(raw_a-raw_b)==4
assert len(raw_m-raw_b)==2
lost,returned=delta(b,m)
assert not lost and len(returned)==2
assert {r['file'] for r in returned}=={'src/compiler/factory/nodeFactory.ts'}
result=dict(mode='void-widening-only-source-scan',before=len(a),after=len(b),mutant=len(m),removed=removed,added=added,mutantReturned=returned)
output.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
