import gzip,hashlib,importlib.util,json,sys
from pathlib import Path
root=Path(__file__).resolve().parent
source=Path('/tmp/any-returns-adapted/src/compiler')
queries=json.loads((root/'queries.json').read_text())
stock=json.loads(gzip.decompress((root/'source-stock.json.gz').read_bytes()))
# The pinned arithmetic is unmodified and has its own historical mutant proofs.
path=Path('/tmp/hidden-census-pin/stage3/census/hidden/hidden.py')
spec=importlib.util.spec_from_file_location('hidden',path);hidden=importlib.util.module_from_spec(spec);spec.loader.exec_module(hidden)
assert hashlib.sha256(path.read_bytes()).hexdigest()==hashlib.sha256((root/'hidden.py.txt').read_bytes()).hexdigest()
assert {p.relative_to(source).as_posix() for p in source.rglob('*') if p.is_file()}==set(stock)
for name,meta in stock.items():
 data=(source/name).read_bytes();assert len(data)==meta['bytes'] and hashlib.sha256(data).hexdigest()==meta['sha256'],name
results=[]
for q in queries:
 row=dict(q);row['sha256']=stock[q['file']]['sha256']
 expected={u['where'] for u in stock[q['file']]['units'] if u['start']<q['end'] and q['start']<u['end']}
 for side in ['main','stack']:
  ledger=root/(side+'-'+q['label']+'.jsonl')
  if not ledger.exists() or not ledger.stat().st_size:continue
  status=root/(side+'-'+q['label']+'.status.json')
  if status.exists() and json.loads(status.read_text())['exit']!=0:
   row[side+'_unmeasured']=json.loads(status.read_text());continue
  rows=[json.loads(line) for line in ledger.read_text().splitlines()]
  assert len(rows)==2 and rows[0]['region_start']==q['start'] and rows[0]['region_end']==q['end']
  assert Path(rows[1]['file'])==source/q['file']
  if '--drop-unit-mutant' in sys.argv:
   rows[1]['units']=rows[1]['units'][1:]
  actual={hidden.local_where(u['where'],source) for u in rows[1]['units']}
  assert actual==expected,(side,q['label'],expected-actual,actual-expected)
  # Never use the whole-file result of this partial ledger; clip to the queried interval.
  calculation=hidden.calculate(rows,{q['file']:stock[q['file']]},source)
  spans=hidden.union((max(a,q['start']),min(b,q['end'])) for a,b in calculation['files'][q['file']]['hidden_ranges'] if a<q['end'] and q['start']<b)
  findings=rows[1]['findings']
  row[side]=dict(hidden_bytes=hidden.size(spans),hidden_ranges=spans,units=sorted(actual),checker_rejected=rows[0]['checker_rejected'],causes=[{k:f.get(k) for k in ['unit','where','kind','reason','start','end']} for f in findings if f['kind'] in ['Refused','NotYet','SkippedDependency','panic']],ledger_sha256=hashlib.sha256(ledger.read_bytes()).hexdigest())
 if 'main' in row and 'stack' in row:
  row['net_revealed_bytes']=row['main']['hidden_bytes']-row['stack']['hidden_bytes']
  row['revealed_ranges']=hidden.subtract(row['main']['hidden_ranges'],row['stack']['hidden_ranges'])
  row['revealed_bytes']=hidden.size(row['revealed_ranges'])
  row['newly_hidden_ranges']=hidden.subtract(row['stack']['hidden_ranges'],row['main']['hidden_ranges'])
  row['newly_hidden_bytes']=hidden.size(row['newly_hidden_ranges'])
  assert row['net_revealed_bytes']==row['revealed_bytes']-row['newly_hidden_bytes']
 results.append(row)
(root/'regions.json').write_text(json.dumps(dict(main='68db8ddd145281a62655452496bdc32ef848bdf3',stack='f742bb83173dcd151abb8f4513985693112c1ea8',source_pin='388096e6a83a4e9d287fb827f793c599ba1bf0ad',adapted_files_verified=len(stock),scope='Ten assigned disjoint hidden intervals only; complete project loaded and registered. Not a whole-corpus total.',regions=results),indent=2)+'\n')
for r in results:print(r['label'],r.get('main',{}).get('hidden_bytes'),r.get('stack',{}).get('hidden_bytes'),r.get('revealed_bytes'))
