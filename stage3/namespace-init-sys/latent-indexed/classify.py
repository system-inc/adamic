"""Classify ledger rows using surviving measurement IR and exact lexical spans."""
import argparse, collections, csv, gzip, hashlib, json, functools
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--raw',type=Path,required=True);p.add_argument('--source',type=Path,required=True);p.add_argument('--output',type=Path,required=True);a=p.parse_args()
repo=Path(__file__).resolve().parents[3];out=a.output;out.mkdir(parents=True,exist_ok=True)
raw=[json.loads(x) for x in a.raw.read_text().splitlines()];header=raw[0]
assert header['project_options'] and header['requires_indexed_presence']
hashes=json.loads((repo/'stage3/stricter-indexed-all/evidence/ledger-source-hashes.json').read_text())
assert len(hashes)==79
for file,expected in hashes.items():assert hashlib.sha256((a.source/file).read_bytes()).hexdigest()==expected,file
files={str(Path(r['file']).relative_to(a.source)):r for r in raw[1:]}
assert len(files)==79
ledger=[r for r in csv.DictReader((repo/'stage3/stricter-indexed-all/evidence/ledger-rows.csv').open()) if r['option']=='noUncheckedIndexedAccess'];assert len(ledger)==99
functions={'FunctionDeclaration','FunctionExpression','ArrowFunction','MethodDeclaration','Constructor','GetAccessor','SetAccessor'}
def field(d,name):return d.get(name,d.get(name[0].upper()+name[1:]))
def sitekey(s):return (str(Path(field(s,'file')).relative_to(a.source)),int(field(s,'line')),int(field(s,'column')),str(field(s,'code')).removeprefix('TS'))
def rowkey(r):return(r['file'],int(r['line']),int(r['column']),r['code'].removeprefix('TS'))
def contains(d,pos):return d['start']<=pos<d['end']
@functools.lru_cache(maxsize=None)
def source_lines(file):
 lines=Path(file).read_text().splitlines(keepends=True);starts=[];position=0
 for line in lines:starts.append(position);position+=len(line.encode())
 return lines,starts
@functools.lru_cache(maxsize=None)
def offset(where):
 file,lineno,column=where.rsplit(':',2);lines,starts=source_lines(file);line=lines[int(lineno)-1];units=0;chars=0
 for ch in line:
  if units>=int(column)-1:break
  units+=len(ch.encode('utf-16-le'))//2;chars+=1
 return starts[int(lineno)-1]+len(line[:chars].encode())
def normalize(f):
 if not f:return None
 f=dict(f)
 for k in ('where','unit','text','reason'):
  if k in f:f[k]=f[k].replace(str(a.source)+'/', '')
 return f
sites={sitekey(field(d,'site')):d for d in header['option_dispositions'] if 'noUncheckedIndexedAccess' in field(field(d,'site'),'options')}
assert set(sites)=={rowkey(r) for r in ledger}
@functools.lru_cache(maxsize=None)
def direct_refusals(file):
 record=files[file];by_owner=collections.defaultdict(list)
 declarations=[d for d in record['declarations'] if d['kind'].removeprefix('Kind') in functions]
 for f in record['findings']:
  if f['phase']!='refusal_scan' or f['kind'] not in ('Refused','NotYet') or not f['where'] or f['where'].rsplit(':',2)[0]!=record['file']:continue
  pos=offset(f['where']);owners=[d for d in declarations if contains(d,pos)]
  if owners:
   closest=min(owners,key=lambda d:d['end']-d['start']);by_owner[(closest['start'],closest['end'])].append(f)
 return by_owner
results=[]
for row in ledger:
 record=files[row['file']];disposition=sites[rowkey(row)];assert field(disposition,'state')=='scheduled-check'
 pos=field(field(disposition,'site'),'position')
 decls=[d for d in record['declarations'] if d['kind'].removeprefix('Kind') in functions and contains(d,pos)]
 owner=min(decls,key=lambda d:d['end']-d['start']) if decls else None
 units=[u for u in record['units'] if contains(u,pos)];assert len(units)==1,(row['id'],units)
 unit=units[0];attempts=[x for x in record.get('function_attempts') or [] if x['where'].rsplit(':',2)[0]==record['file'] and owner and x['start']==owner['start'] and x['end']==owner['end'] and x['unit']==unit['where']]
 checks={(c['unit'],c['where']) for c in record.get('checks') or [] if c['kind']=='indexed-presence'}
 emitted=[r for r in record.get('reads') or [] if (r['unit'],r['where']) in checks and rowkey(row) in {sitekey(s) for s in r['sites']}]
 own_refusals=direct_refusals(row['file']).get((owner['start'],owner['end']),[]) if owner else []
 failed=[x['failure'] for x in attempts if x.get('failure')]
 blocker=next((f for f in failed if f['kind'] in ('Refused','NotYet')),None) or (own_refusals[0] if own_refusals else None)
 scope='function';why='';state=''
 if blocker:state='function blocked by a refusal';why='actual lowering failure' if blocker in failed else 'direct syntax refusal in the row function'
 elif emitted:state='check emitted';why='panic-bearing indexed-presence check survives in measurement IR'
 else:
  state='not reached';scope='enclosing attempt'
  enclosing=[x for x in record.get('function_attempts') or [] if x['where'].rsplit(':',2)[0]==record['file'] and contains(x,pos) and x.get('failure')]
  if enclosing:blocker=min(enclosing,key=lambda x:x['end']-x['start'])['failure'];why='enclosing function attempt stops before a surviving row check'
  elif unit['status']=='skipped_checker_body':why='enclosing top-level function skipped for checker body diagnostics: '+'; '.join(unit['checker_diagnostics'])
  else:
   fs=[f for f in record['findings'] if f['unit']==unit['where'] and f['phase']!='refusal_scan']
   blocker=fs[0] if fs else (failed[0] if failed else None)
   why='enclosing top-level attempt stopped' if blocker else 'no matching guard survived the isolated top-level attempt; no instantiated nested context was invented'
 results.append(dict(id=row['id'],file=row['file'],line=int(row['line']),column=int(row['column']),ledger_owner=row['owner'],owner=normalize(owner),unit=normalize(unit),state=state,why=why,blocker_scope=scope if blocker else '',blocker=normalize(blocker),checks=[normalize(r) for r in emitted]))
assert len(results)==99 and len({r['id'] for r in results})==99
counts=collections.Counter(r['state'] for r in results)
blocks=collections.defaultdict(list)
for r in results:
 b=r['blocker']
 if b and b['kind'] in ('Refused','NotYet'):blocks[(b['kind'],b['where'],b['text'])].append(r)
table=[dict(kind=k[0],where=k[1],message=k[2],rows=len(rs),ids=[r['id'] for r in rs],scopes=dict(collections.Counter(r['blocker_scope'] for r in rs))) for k,rs in blocks.items()];table.sort(key=lambda r:(-r['rows'],r['where']))
summary=dict(compiler='5ad36d2cf4fce475ec072f3cc68e21c393dffc45',measurement_only=True,source_files=79,source_hash_mismatches=0,indexed_rows=99,counts=dict(counts),blocking_refusals=table,option_dispositions=collections.Counter(field(d,'state') for d in header['option_dispositions']))
(out/'rows.json').write_text(json.dumps(results,indent=2)+'\n');(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
with (out/'rows.csv').open('w') as f:
 w=csv.DictWriter(f,lineterminator='\n',fieldnames=['id','file','line','column','ledger_owner','state','why','blocker_scope','blocker_location','blocker_message']);w.writeheader()
 for r in results:w.writerow({**{k:r[k] for k in w.fieldnames if k in r},'blocker_location':(r['blocker'] or {}).get('where',''),'blocker_message':(r['blocker'] or {}).get('text','')})
with gzip.open(out/'raw.jsonl.gz','wb') as f:f.write(a.raw.read_bytes())
(out/'source-identity.json').write_text(json.dumps(hashes,indent=2)+'\n')
report=['Measurement at `5ad36d2c`; all production refusals retained.','',str(dict(counts)),'','Checks mean surviving panic-bearing IR in an isolated latent attempt, not native output. Nested functions are not retried with fabricated lexical environments. Direct syntax refusals name the smallest lexical function; propagated failures retain their actual source location. The one primary blocker per row conserves row counts. The raw census includes additional refusals beyond this attribution.','','| Rows | Own function | Enclosing stop | Location | Kind | Message |','| ---: | ---: | ---: | --- | --- | --- |']
for b in table:report.append(f"| {b['rows']} | {b['scopes'].get('function',0)} | {b['scopes'].get('enclosing attempt',0)} | {b['where']} | {b['kind']} | {b['message'].replace('|', '&#124;').replace(chr(10),' ')} |")
report+=['','See rows.csv for all 99 row classifications, including reasons for not reached.','Production project options retained; scheduled contracts are not checker-body skips. Remaining checker errors are retained and skip only affected bodies.','Generic declarations are attempted as written. Module ordering, backend emission, and final ownership passes are outside this measurement. No backend runs: the measurement binary disables production Lower and ordinary Load.']
(out/'REPORT.md').write_text('\n'.join(report)+'\n');print(json.dumps(summary,indent=2))
