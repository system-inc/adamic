import pathlib,json,statistics,re,shutil
p=pathlib.Path('review/test-audit/internal-lower-non_null_impossible');scope=json.loads((p/'scope.json').read_text());names=scope['rows'];family=scope['family'];items=json.loads((p/'manifest.json').read_text());cmds=json.loads((p/'matrix-commands.json').read_text());bylabel={c['label']:c for c in cmds}
def events(label):
 f=p/(label+'.log');out=[]
 if not f.exists():return out
 for l in f.read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
def row(n):return names[-1] if n in family else n
def members(n):return family if n==names[-1] else [n]
def median(n):return statistics.median(next(e['Elapsed'] for e in events(n.replace(' ','_')+'.'+str(i)) if e.get('Action')=='pass' and not e.get('Test')) for i in range(1,4))
def diagnostic(es,n):
 return next((e['Output'].strip() for e in es if e.get('OutputType')=='error' and e.get('Test','').split('/')[0] in members(n)), next((e['Output'].strip() for e in es if 'panic:' in e.get('Output','')), 'No diagnostic extracted'))
mat=[];kills={n:[] for n in names};unique={n:[] for n in names};probes={n:[] for n in names};proof={};probe_status={n:None for n in names};allkills={};bounded=set()
for it in items:
 id=it['id'];labels=[c['label'] for c in cmds if c['id']==id];narrow=id.startswith('M') and any(l!=id for l in labels);es=events(id);fail=set();obs={}
 if id.startswith('M') and not narrow:
  fail={row(e['Test']) for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']}
  for n in names:
   statuses=[e['Action'] for e in es if e.get('Test') in members(n) and e.get('Action') in ['pass','fail']]
   obs[n]='fail' if n in fail else ('pass' if len(statuses)==len(members(n)) and all(s=='pass' for s in statuses) else 'unknown')
 else:
  for label in labels:
   if label==id:continue
   n=next((n for n in names if label==id+'.'+n.replace(' ','_')),None)
   if n is None:continue
   one=events(label);statuses=[e['Action'] for e in one if e.get('Test') in members(n) and e.get('Action') in ['pass','fail']]
   if 'fail' in statuses or (bylabel[label]['exit']!=0 and any(e.get('Test','').split('/')[0] in members(n) for e in one)):status='fail'
   elif len(statuses)==len(members(n)) and all(s=='pass' for s in statuses):status='pass'
   else:status='unknown'
   obs[n]=status
   if status=='fail':fail.add(n)
  if narrow:bounded.update(names)
 for n in names:
  if obs.get(n)=='fail':
   if id.startswith('P'):probes[n].append(id)
   else:
    kills[n].append(id);label=id+'.'+n.replace(' ','_') if narrow else id;proof[n]=(id,label,diagnostic(events(label),n))
  if id.startswith('P') and n in obs:probe_status[n]=obs[n]
 if id.startswith('M') and len(fail)==1:
  n=next(iter(fail))
  if n in unique:unique[n].append(id)
 for n in fail:allkills.setdefault(n,[]).append(id)
 mat.append(dict(id=id,bounded=narrow,failed_rows=sorted(fail),observations=obs,matrix_rows=names if narrow else 'whole package',survivor=id.startswith('M') and not fail,outside_scope='unknown' if narrow else 'observed'))
(p/'matrix.json').write_text(json.dumps(mat,indent=2)+'\n')
oracles=[
'Handwritten Refused class, exact non-null What and Fix for nullish .a assertions.',
'Handwritten IR Coalesce/Panic presence, Number/MaybeNumber types and panic-message prefix/suffix.',
'Handwritten NumberConstant type and zero panic strings; does not check the numeric value. M04 changes 0 to 1 and passes.',
'Handwritten NotYet class and exact structural-Map diagnostic.',
'Handwritten Refused class, exact index-signature What and Fix.',
'Handwritten diagnostic classes/text for four negative cases; last continuation case checks only err == nil.',
'Only successful census construction/encoding; no expected site count or contents. Fixture-derived project: baseline one site; empty-answer probe zero sites still passes.',
'Handwritten Refused diagnostics and own .refused snapshots; incompatible subcase checks a preparation-time TypeScript checker error.',
'Only err == nil for class/fresh/declared cases; no IR values or execution assertion.',
'Only refuse returns nil for overwritten spread; no positive refusal case.',
'Handwritten inferred type-map entries/absence for four binder cases; ordinary production test despite Witnesses name.',
'Handwritten NotYet class and parameter/fix substrings, one checker over three fixtures.'
]
secs={n:median(n) for n in names};secs['TestInheritanceNativeSignatureLimits']=statistics.median(next(e['Elapsed'] for e in events('subsumer.'+str(i)) if e.get('Action')=='pass' and not e.get('Test')) for i in range(1,4));rows=[]
for i,n in enumerate(names):
 file=None;line=None
 for f in pathlib.Path('internal/lower').glob('*_test.go'):
  for k,l in enumerate(f.read_text().splitlines(),1):
   if l.startswith('func '+members(n)[0]+'('):file=str(f);line=k
 if unique[n]:verdict='slow-worthy' if secs[n]>60 else 'sacred';sub=[]
 elif not kills[n]:verdict='untrue';sub=[]
 else:
  candidates=[o for o in allkills if o!=n and set(kills[n])<=set(allkills[o]) and not o.startswith('P')]
  chosen=('TestInheritanceNativeSignatureLimits' if n==names[-1] and 'TestInheritanceNativeSignatureLimits' in candidates else min(candidates,key=lambda o:secs.get(o,999))) if candidates else None
  if chosen:verdict='subsumed';sub=[chosen]
  else:verdict='overlapping';sub=sorted({o for o in allkills if o!=n and set(kills[n])&set(allkills[o])})
 last=None;evidence='No production mutation caught this row; matrix.json contains every observed result.'
 if n in proof:
  id,label,diag=proof[n];last=id+' '+diag;evidence=bylabel[label]['command']+'; '+diag
 own=probe_status[n];obj=dict(test=n,package='internal/lower',file=file+':'+str(line),seconds=secs[n],oracle=oracles[i],oracle_kind='self',kills=kills[n],unique_kills=unique[n],last_proven_fail=last,verdict=verdict,subsumed_by=sub,mutants_in_matrix=16,probe_kills=probes[n],subsumer_seconds=secs.get(sub[0]) if verdict=='subsumed' else None,vacuous=True if own=='pass' else False if own=='fail' else None,bounded=n in bounded,matrix_rows=names if n in bounded else [],evidence=evidence)
 if n==names[-1]:obj['members']=family
 if verdict=='subsumed':obj['subsumption_mutants']=len(kills[n])
 if n=='TestOptionalIndexingKeepsUnsupportedStorageNotYet':obj['vacuous_subcases']=['index_after_ordinary_continuation']
 if n=='TestOptionalWideningRefused':obj['preparation_only_subcases']=['incompatible']
 rows.append(obj)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
md=['| ID | Origin file:line | Change | Failed rows |','|---|---|---|---|']
for it,m in zip(items,mat):
 if it['id'].startswith('P'):continue
 md.append('| '+it['id']+' | '+it['file']+':'+str(it['line'])+' | '+(it['old']+' -> '+it['new']).replace('|','\\|')+' | '+', '.join(m['failed_rows'])+(' (bounded)' if m['bounded'] else '')+' |')
(p/'mutants.md').write_text('\n'.join(md)+'\n')
for src in ['plan','matrix','baselines','vet','report']:shutil.copyfile('/tmp/u040-'+src+'.py',p/(src+'.py'))
print(json.dumps([(r['test'],r['verdict'],r['kills'],r['unique_kills'],r['subsumed_by'],r['vacuous']) for r in rows],indent=2))
