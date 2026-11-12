"""Pair completed boundary observations and retain unit, raw-site and distinct-read counts."""
import collections,gzip,hashlib,json,sys
from pathlib import Path
D=Path(__file__).resolve().parent;repo=D.parents[2];sys.path.insert(0,str(repo/'stage3/census/pair'));import pair
TREE=Path('/workspace/slice3-adapted');TS=repo/'stage3/api/node_modules/typescript'
def load(p):
 if not p.exists():p=Path(str(p)+'.gz')
 with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as f:return [json.loads(s) for s in f if s.strip()]
def save(name,data):(D/name).write_text(json.dumps(data,indent=2)+'\n')
def isview(r):return r.get('view_stop') is True
obs={};normal={};converted={};byunit={};ordinary={};summary={}
for side in ['base','head']:
 rows=[];head=None;records=[]
 for i,selection in enumerate(json.loads((D/'groups.json').read_text())):
  files=selection['files']
  p=D/f'raw/group-{i}-{side}';meta=json.loads(p.with_suffix('.json').read_text());assert meta['child_exit']==0 and not meta['timed_out'],meta
  current=load(Path(str(p)+'-normal.jsonl'));assert current[0]['status']=='measurement' and not current[0]['checker_rejected'],current[0]
  if head is None:head=current[0]
  else:assert head==current[0]
  file_rows=[r for r in current[1:] if 'units' in r];assert {r['file'] for r in file_rows}==set(files),(side,i,'selection coverage')
  rows+=file_rows;records+=load(Path(str(p)+'-records.jsonl'))
 merged={}
 for r in rows:
  file=r['file']
  if file not in merged:merged[file]=dict(r,units=[],findings=[])
  m=merged[file];known={u['where'] for u in m['units']};assert not (known & {u['where'] for u in r['units']}),(side,file,'overlapping unit ranges')
  m['units']+=r['units']
  scans=[f for f in r.get('findings') or [] if f.get('phase')=='refusal_scan']
  previous=[f for f in m['findings'] if f.get('phase')=='refusal_scan']
  if previous:assert previous==scans,(side,file,'refusal scan changed across ranges')
  else:m['findings']+=scans
  m['findings'] += [f for f in r.get('findings') or [] if f.get('phase')!='refusal_scan']
 rows=list(merged.values())
 assert len(rows)==len(head['sources'])
 normal[side]=[head]+sorted(rows,key=lambda r:r['file']);obs[side]=records;errors=[r for r in records if r['kind']=='BoundaryProbeError'];assert not errors,errors
 byunit[side]={r['unit']:r for r in records};assert len(byunit[side])==len(records)
 attempted={u['where'] for r in rows for u in r['units'] if u['status']=='attempted'};assert attempted<=set(byunit[side]),attempted-set(byunit[side])
 ordinary[side]=collections.defaultdict(list)
 for r in rows:
  for f in r.get('findings') or []:
   if f.get('phase') in ('lowering','deferred'):ordinary[side][f['unit']].append(f)
 grouped=collections.defaultdict(list)
 for r in records:
  if not isview(r):continue
  grouped[r['file']].append({'kind':'NotYet','measurement':head['measurement'],'phase':'lowering','unit':r['unit'],'where':r['where'],'reason':r['reason'],'text':r.get('text','')})
 converted[side]=[head]+[dict(r,findings=grouped[r['file']]) for r in normal[side][1:]]
 (D/f'{side}-for-pair.jsonl').write_text(''.join(json.dumps(r,separators=(',',':'))+'\n' for r in converted[side]))
 summary[side]={'observed_units':len(records),'views_recorded':sum(r['views'] for r in records),'boundary_stopped_units':sum(r.get('outcome')=='stop' for r in records),'view_stopped_units':sum(isview(r) for r in records),'other_boundary_stopped_units':sum(r.get('outcome')=='stop' and not isview(r) for r in records),'native_host_walk_calls':sum(r.get('real_host_walk_calls',0) for r in records),'ordinary_unit_status':dict(collections.Counter(u['status'] for r in rows for u in r['units'])),'ordinary_finding_observations':sum(len(r.get('findings') or []) for r in rows),'ordinary_panics':sum(f['kind']=='panic' for r in rows for f in r.get('findings') or [])}
verification={}
for i in range(len(json.loads((D/'groups.json').read_text()))):
 def ordinary_bytes(side):
  p=D/f'raw/group-{i}-{side}-normal.jsonl'
  if not p.exists():p=Path(str(p)+'.gz')
  raw=gzip.open(p,'rb').read() if p.suffix=='.gz' else p.read_bytes()
  return b''.join(line for line in raw.splitlines(keepends=True) if json.loads(line).get('status')!='relation_report')
 a=ordinary_bytes('base');b=ordinary_bytes('head');assert a==b,(i,'ordinary non-telemetry stream bytes changed')
 verification[str(i)]={'ordinary_non_telemetry_bytes_identical':True,'sha256':hashlib.sha256(a).hexdigest()}
save('ordinary-byte-verification.json',verification)
assert byunit['base'].keys()==byunit['head'].keys(),'unit observer coverage differs'
assert {k:r['views'] for k,r in byunit['base'].items()}=={k:r['views'] for k,r in byunit['head'].items()},'per-unit view-source populations differ'
summary['per_unit_view_counts_equal']=True
assert summary['head']['native_host_walk_calls']>0,'changed walker never ran'
delta=pair.compare(converted['base'],converted['head'],TREE,TS);save('view-delta.json',delta);(D/'view-delta.md').write_text(pair.report(delta))
normaldelta=pair.compare(normal['base'],normal['head'],TREE,TS);save('ordinary-delta.json',normaldelta)
ids=pair.identities([normal['base'],converted['base'],converted['head']],TREE,TS)
new=[];cleared=[];moved=[]
def counter_view(r):return r.get('stop') and (r.get('reason','').startswith(('a view of ','a view as ')))
for unit,h in byunit['head'].items():
 b=byunit['base'][unit]
 if isview(b) and not isview(h):cleared.append({'unit':unit,'base':b,'head':h})
 if isview(b) and isview(h) and (b['where'],b['reason'])!=(h['where'],h['reason']):moved.append({'unit':unit,'base':b,'head':h})
 if not isview(h) or isview(b):continue
 assert h.get('text'),h
 c=h['counterfactuals'];a=not counter_view(c['without_deeper']);q=not counter_view(c['without_unconstrained']);both=not counter_view(c['without_both'])
 cause='a deeper host walk' if a and not q else 'b unconstrained parameter' if q and not a else 'a and b both necessary' if a and q else 'a and b independently sufficient' if both else 'neither widening removes the unit view stop'
 entry={'id':len(new)+1,'unit':unit,'head':h,'base_boundary':b,'base_ordinary_findings':ordinary['base'].get(unit,[]),'base_no_other_finding':not ordinary['base'].get(unit) and b.get('outcome')=='clear','cause':cause,'without_deeper_removes_unit_view_stop':a,'without_unconstrained_removes_unit_view_stop':q,'without_both_removes_unit_view_stop':both};new.append(entry)
save('new-stopped-units.json',new);save('cleared-units.json',cleared);save('moved-units.json',moved)
def tally(entries):
 raw=set();distinct=set()
 for e in entries:
  f=e['head'];i=ids[f['where']];raw.add((i['file'],i['sha256'],i['byte_offset'],f['reason'],f['text']));distinct.add((i['file'],i['sha256'],i['read_offset'],f['reason']))
 return {'unit_observations':len(entries),'raw_findings':len(raw),'distinct_reads':len(distinct)}
summary['new_stops']=tally(new);summary['base_finding_free_regressions']=tally([e for e in new if e['base_no_other_finding']]);summary['causes']={cause:tally([e for e in new if e['cause']==cause]) for cause in ['a deeper host walk','b unconstrained parameter','a and b both necessary','a and b independently sufficient','neither widening removes the unit view stop']};summary['cleared_view_stopped_units']=len(cleared);summary['moved_view_first_stop_units']=len(moved)
summary['ordinary_pair']={k:{x:len(normaldelta[k][x]) for x in ['lowered','moved','new_findings','unit_first_stop_delta']} for k in ['raw_findings','distinct_reads']};summary['new_panics']=[r for r in normaldelta['raw_findings']['regressions'] if r['type']=='new_panic']
all_cleared=[];all_moved=[];all_new=[]
for unit,h in byunit['head'].items():
 b=byunit['base'][unit];bs=b.get('outcome')=='stop';hs=h.get('outcome')=='stop'
 if bs and not hs:all_cleared.append({'unit':unit,'base':b,'head':h})
 if bs and hs and (b['where'],b['reason'])!=(h['where'],h['reason']):all_moved.append({'unit':unit,'base':b,'head':h})
 if hs and not bs:all_new.append({'unit':unit,'base':b,'head':h})
save('all-boundary-cleared-units.json',all_cleared);save('all-boundary-moved-units.json',all_moved);save('all-boundary-new-stopped-units.json',all_new)
summary['all_boundary_unit_changes']={'cleared':len(all_cleared),'moved':len(all_moved),'new':len(all_new)}
summary['view_pair_counts']={k:{x:len(delta[k][x]) for x in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta']} for k in ['raw_findings','distinct_reads']}
for side in ['base','head']:
 summary[side]['raw_view_findings']=sum(delta['raw_findings']['totals']['before' if side=='base' else 'after'].values());summary[side]['distinct_view_reads']=sum(delta['distinct_reads']['totals']['before' if side=='base' else 'after'].values())
summary['all_ordinary_file_unit_findings_equal']=normal['base']==normal['head'];save('summary.json',summary)
print(json.dumps(summary,indent=2))
