import pathlib,json,re,statistics,itertools
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-native-arguments_length');names=json.loads((p/'names.json').read_text());times=json.loads((p/'timings.json').read_text());runs=json.loads((p/'runs.json').read_text());wit={'TestClosureConventionDropCount':'W_BUILD','TestClosureConventionRuntimeDropCount':'W_BUILD','TestClosureConventionWrongOrder':'W_BUILD','TestOptionalMethodThunksMatchNode':'W_NODE'};matrix={};lines={};probes=json.loads((p/'probes.json').read_text());owners={n:[] for n in names}
for id,file,old,ret,rows in probes:
 for n in rows:owners[n].append(id)
for run in runs:
 id=run['id'];fails=set();seen=set();passthrough=[];lineby={};logs=[p/(id+'.log')]
 if run['panic'] or run['cooked']:logs += list(p.glob(id+'-Test*.log'))
 for log in logs:
  raw=log.read_text()
  if log.name!=id+'.log' and 'panic:' in raw:
   n=log.name[len(id)+1:-4];fails.add(n);seen.add(n)
  for l in raw.splitlines():
   try:d=json.loads(l)
   except:continue
   t=d.get('Test','');top=t.split('/')[0]
   if d.get('Action') in ['fail','pass','skip'] and t:seen.add(top)
   if d.get('Action')=='fail' and t:fails.add(top)
   if d.get('Action')=='pass' and '/' in t:passthrough.append(t)
   output=d.get('Output','').strip()
   if t and re.search(r'_test.go:\d+:',output):lineby.setdefault(top,output)
   if output.startswith('Node ') and top in lineby and ': native ' in lineby[top] and '; Node ' not in lineby[top]:lineby[top]+='; '+output
 matrix[id]={'failed':sorted(fails),'eligible_failed':sorted(fails-set(wit)) if id.startswith('M') else sorted(fails),'witness_precondition_failures':sorted(fails&set(wit)) if id.startswith('M') else [],'seen':sorted(seen),'passing_subcases':passthrough,'code':run['code'],'wall':run['wall'],'rows':run['rows']};lines[id]=lineby
rows=[]
for i,n in enumerate(names,1):
 kills=[id for id,m in matrix.items() if id.startswith('M') and n in m['eligible_failed']];unique=[id for id in kills if matrix[id]['eligible_failed']==[n]];sub=[];w=wit.get(n)
 if w:verdict='witness' if n in matrix.get(w,{}).get('failed',[]) else 'untrue'
 elif unique:verdict='sacred'
 elif not kills:verdict='untrue'
 else:
  common=set.intersection(*(set(matrix[id]['eligible_failed'])-{n} for id in kills))
  if common:verdict='subsumed';sub=[min(common,key=lambda x:statistics.median(times[x]))]
  else:
   verdict='overlapping';candidates=sorted(set().union(*(set(matrix[id]['eligible_failed'])-{n} for id in kills)))
   for size in range(2,len(candidates)+1):
    choices=[c for c in itertools.combinations(candidates,size) if all(any(x in matrix[id]['eligible_failed'] for x in c) for id in kills)]
    if choices:sub=list(choices[0]);break
 owned=owners[n];pk=[id for id in owned if n in matrix.get(id,{}).get('failed',[])];vacuous_entries=[id for id in owned if n in matrix.get(id,{}).get('seen',[]) and id not in pk]
 chosen=w if w else unique[-1] if unique else kills[-1] if kills else None;line=lines.get(chosen,{}).get(n)
 oracle='Handwritten production-plan or emitted-C assertions';kind='self'
 if n=='TestCaseMappingMatchesNode':oracle='Native WTF-8 mapping compared byte-for-byte with Node toUpperCase/toLowerCase; handwritten sweep line counts';kind=['external-run','self']
 if n=='TestCaseTablesMatchNodesUnicode':oracle='Node process.versions.unicode compared with production Unicode metadata';kind='external-run'
 if n=='TestArithmeticIsNeverFused':oracle='Handwritten 0 0 0 output; separately built clang fast-contraction control proves FMA is observable'
 if n in wit:
  oracle='Clang rejects planted typed-ABI violations; handwritten diagnostic substrings' if w=='W_BUILD' else 'Node/native agreement plus handwritten convention/thunk checks; planted missing-thunk disagreement witness'
  kind=['external-run','self']
 if n=='TestInheritanceMemoryPlans':oracle+='; empty element-borrow plan passes while empty region and reuse plans fail'
 if n=='TestParserHasNoUnusedOptionalMethodThunks':oracle='Handwritten absence of three optional-thunk declarations and IR fixture guards; empty C output passes'
 if n=='TestClosureConventionRuntimeFeaturesIgnoreLiterals':oracle='Handwritten absence of unused feature macros in built runtime header; empty emitted C passes'
 file=next(f for f in ['arguments_length_test.go','borrow_chain_test.go','borrow_consumes_test.go','case_test.go','class_inheritance_test.go','closure_convention_test.go','contract_test.go'] if ('func '+n+'(') in pathlib.Path('/workspace/adamic/internal/native',f).read_text())
 cmd=('ADAMIC_MUTANT='+chosen+' ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/'+chosen+' timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '+repr('^('+'|'.join(matrix[chosen]['rows'])+')$')+' > '+chosen+'.log 2>&1') if chosen else None
 r=dict(row_id='R'+str(i),test=n,package='internal/native',file='internal/native/'+file,seconds=statistics.median(times[n]),oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=unique,last_proven_fail=(chosen+': '+line) if line else None,verdict=verdict,subsumed_by=sub,mutants_in_matrix=20,probe_kills=pk,subsumer_seconds=statistics.median(times[sub[0]]) if verdict=='subsumed' else None,vacuous=bool(vacuous_entries) if owned and all(n in matrix.get(id,{}).get('seen',[]) for id in owned) else None,bounded=True,matrix_rows=names,evidence=(cmd+'; '+line) if line else 'No eligible failure observed')
 if vacuous_entries:r['vacuous_entries']=[{'P_C':'C','P_BORROW':'planElementBorrows'}.get(id,id) for id in vacuous_entries]
 if w:r['witness_check']=w
 pass_subs=[s for id in owned for s in matrix.get(id,{}).get('passing_subcases',[]) if s.startswith(n+'/')]
 if pk and pass_subs:r['vacuous_subcases']=pass_subs
 rows.append(r)
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));(p/'rows.json').write_text(json.dumps(rows,indent=2));print([(r['test'],r['verdict'],r['kills'],r['unique_kills'],r['vacuous'],r.get('vacuous_entries')) for r in rows])
