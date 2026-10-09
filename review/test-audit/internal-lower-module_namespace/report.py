import pathlib,json,re,statistics,subprocess,gzip,hashlib
p=pathlib.Path('review/test-audit/internal-lower-module_namespace');rows=json.loads((p/'rows.json').read_text());matrix=json.loads((p/'matrix.json').read_text());probes=json.loads((p/'probe-results.json').read_text());plan=json.loads((p/'mutant-plan.json').read_text());names=[r['test'] for r in rows]
oracles=[
'Self-written NotYet type and diagnostic substrings; export-write subcase only requires Load to return any error and never calls Lower.',
'Self-written absence of one ReferenceError string in emitted C; does not require the value read or output to exist.',
'Self-written NotYet type and reason substrings for six unsupported observations.',
'Self-written Refused type and different receivers substring.',
'Self-written NotYet type and namespace initialization substring.',
'Self-written Refused type and different receivers substring.',
'Self-written nil preflight expectation and host error text; cwd also accepts an independent cycle-capable Refused, so another stop may satisfy it.',
'Self-written booleans for ambient, executable, and declaration-file namespace contexts.',
'Self-written exact discovery counter values, not measured asymptotic work; counts can change without traversal changing.',
'Self-written exact namespace identity set and discovery count for one three-node cycle.',
'Self-written NotYet type and enum initialization substring.',
'Self-written nil errors from declaration and body lowering; returned statements and bindings are not asserted.',
'Self-written NotYet type, exact program location, and missing-symbol substring.',
'Self-written success and diagnostic expectations for normalized TypeScript shapes; overload expectation is derived from Adamic itself, not tsc.'
]
def fail_line(m,n):
 events=m.get('events',[])
 if m.get('bounded'):
  log=p/'logs'/(m['id']+'-'+n+'.log')
  if log.exists():
   events=[]
   for s in log.read_text().splitlines():
    if s.startswith('{'):
     try:events.append(json.loads(s))
     except:pass
 candidates=[e.get('Output','').strip() for e in events if e.get('Test','').split('/')[0]==n and e.get('Action')=='output' and re.search(r'_test.go:\d+:|^panic:',e.get('Output',''))]
 if candidates:return candidates[-1]
 return next((e.get('Output','').strip() for e in events if e.get('Output','').startswith('panic:')),'failure, see log')
report=[]
for index,r in enumerate(rows):
 n=r['test'];kills=[m['id'] for m in matrix if n in m['kills']];unique=[m['id'] for m in matrix if m['kills']==[n]]
 candidates=set.intersection(*(set(m['kills'])-{n} for m in matrix if n in m['kills'])) if kills else set()
 bounded=any(m.get('bounded') for m in matrix)
 subs=[];subs_seconds=None
 if unique:verdict='sacred'
 elif kills and candidates:
  # Choose the fastest measured subsumer. Unmeasured outside rows require timing before finalizing.
  measured={q['test']:q['seconds'] for q in rows}
  extra=p/'subsumer-times.json'
  if extra.exists():measured.update(json.loads(extra.read_text()))
  chosen=min(candidates,key=lambda x:(measured.get(x,float('inf')),x));subs=[chosen];subs_seconds=measured.get(chosen);verdict='subsumed'
 elif kills:verdict='overlapping';subs=sorted(set.union(*(set(m['kills'])-{n} for m in matrix if n in m['kills'])))
 else:verdict='untrue'
 own=[q for q in probes if q['test']==n];pk=[q['id'] for q in own if q['exit']!=0];empty_pass={q['id']:q['exit']==0 for q in own}
 vacuous=all(empty_pass.values()) if own else None
 if index==11:vacuous=empty_pass.get('P08')
 latest=next((m for m in reversed(matrix) if n in m['kills']),None)
 evidence=None if latest is None else latest['command']+'; ADAMIC_MUTANT='+latest['id']+'; '+fail_line(latest,n)
 result=dict(test=n,package='internal/lower',file=r['file']+':'+str(r['line']),seconds=r['seconds'],oracle=oracles[index],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=None if latest is None else latest['id']+': '+fail_line(latest,n),verdict=verdict,subsumed_by=subs,mutants_in_matrix=len(matrix),probe_kills=pk,subsumer_seconds=subs_seconds,vacuous=vacuous,bounded=bounded,matrix_rows=names if bounded else [],evidence=evidence,entry_probe_passes=empty_pass)
 if index==0:result['unprobed_subcases']=['performance.value = 2; is a Load/checker assertion outside lowering']
 if index==7:result['vacuous_subcases']=['ambient namespace','inherited ambient namespace','declaration-file namespace']
 if index==13:result['vacuous_subcases']=['module overload policy','BuilderState','JsxNames','ReactNames','BinaryExpressionState','Parser.JSDocParser','Debug','Debug.log','Parser','IncrementalParser']
 if verdict=='subsumed':result['subsumption_basis']=str(len(kills))+' production mutants; hint, not deletion'
 if verdict=='untrue':result['limitation']='No production mutant in this fixed 15-mutant menu made this row fail; this finding is limited to the planted menu.'
 report.append(result)
(p/'report.json').write_text(json.dumps(report,indent=2))
print(json.dumps(report,indent=2))
print('MUTANTS')
for m in matrix:print(m['id'],m['kills'])
