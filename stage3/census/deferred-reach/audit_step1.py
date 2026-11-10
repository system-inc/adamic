"""Audit retained Step 1 counts against complete immutable raw streams."""
import gzip,hashlib,json,sys
from pathlib import Path
root=Path(sys.argv[1]).resolve()
manifest=json.loads((root/'evidence-sha256.json').read_text())
for name,want in manifest.items():
 assert hashlib.sha256((root/'evidence'/name).read_bytes()).hexdigest()==want,name
report=json.loads((root/'STEP1.json').read_text())
for name,scope in report['scopes'].items():
 raw=gzip.decompress((root/'evidence'/(name+'.jsonl.gz')).read_bytes())
 assert hashlib.sha256(raw).hexdigest()==scope['raw_sha256']
 rows=[json.loads(x) for x in raw.splitlines()];assert rows[-1]['status']=='relation_report'
 files=[r for r in rows[1:] if 'file' in r];actual={};units={}
 for record in files:
  file='src/'+record['file'].split('/src/',1)[1]
  deferred=[u for u in record['units'] if u['status']=='deferred_generic_instantiation']
  actual[file]=dict(file=file,units=len(record['units']),instantiated_true=sum(u['instantiated'] is True for u in deferred),instantiated_false=sum(u['instantiated'] is False for u in deferred))
  for u in deferred:
   where='src/'+u['where'].split('/src/',1)[1];assert where not in units;units[where]=u
 assert actual=={r['file']:r for r in scope['files']}
 assert len(units)==len(scope['deferred_units'])
 for unit in scope['deferred_units']:assert unit['instantiated'] is units[unit['where']]['instantiated']
 print(name,len(files),'files',len(units),'deferred units',sum(u['instantiated'] for u in units.values()),'reached')
 assert scope['files'][0]['file']=='src/compiler/core.ts'
 if name=='base':
  assert len(files)==79 and len(units)==504
  core=actual['src/compiler/core.ts'];assert core['units']==264 and core['instantiated_true']+core['instantiated_false']==207
stock=json.loads(gzip.decompress((root/'evidence/generic-declarations.json.gz').read_bytes()))
assert len(stock)==504 and sum(not u['has_body'] for u in stock)==137
assert sum(not u['has_body'] and u['where'].startswith('src/compiler/core.ts:') for u in stock)==84
print('PASS: hashes, complete trailers, per-file counts, per-unit flags, and independent bodyless inventory')
