import pathlib,json,time,subprocess,os,re,statistics,difflib,collections,datetime,shutil
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/internal-lower-predicates_overload';P=R/'internal/lower';rows=json.load(open(E/'scope.json'));plans=json.load(open(E/'plan.json'))
def events(path):
 out=[]
 for l in path.read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
def terminal(es):return [e for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']]
def failure(path,row):
 for e in events(path):
  if e.get('Test','').split('/')[0]==row and e.get('Action')=='output' and re.search(r'_test.go:\d+:',e.get('Output','')):return e['Output'].strip()
 for e in events(path):
  if 'panic: runtime error' in e.get('Output',''):return e['Output'].strip()
 return None
while not (E/'witness.done').exists():time.sleep(1)
baseline=events(E/'baseline.log');count=len(terminal(baseline));matrix={};bounds=[];checks={};logs={}
for p in plans:
 id=p['id'];es=events(E/(id+'.log'));ts=terminal(es);checks[id]={'completed_top_level':len(ts),'expected_top_level':count,'skips':[e['Test'] for e in ts if e['Action']=='skip']};fails={e['Test'] for e in ts if e['Action']=='fail'}
 if len(ts)<count:
  bounds.append(id);fails=set()
  for row in rows:
   path=E/(id+'-'+row+'.log');ee=events(path);pkg=[e for e in ee if not e.get('Test') and e.get('Action') in ['pass','fail']];assert pkg,(id,row)
   if pkg[-1]['Action']=='fail':fails.add(row)
   logs[id,row]=path
 else:
  for row in rows:logs[id,row]=E/(id+'.log')
 matrix[id]=sorted(fails)
(E/'matrix.json').write_text(json.dumps(matrix,indent=2));(E/'completion-checks.json').write_text(json.dumps(checks,indent=2));(E/'bounded-mutants.json').write_text(json.dumps(bounds))
med={}
for row in rows:
 vals=[]
 for i in range(3):vals.append(next(e['Elapsed'] for e in reversed(events(E/f'timing-{row}-{i}.log')) if not e.get('Test') and e.get('Action')=='pass'))
 med[row]=round(statistics.median(vals),3)
pmat={};vsub={}
for p in json.load(open(E/'probe-time.json')):
 key=p['id'];row=p['row'];pmat.setdefault(key,[])
 if p['exit']!=0:pmat[key].append(row)
 es=events(E/(key+'-'+row+'.log'));vsub[row]=[e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test','').startswith(row+'/')]
(E/'probe-matrix.json').write_text(json.dumps(pmat,indent=2))
result=[];allrows=[x for x in (E/'test-list.log').read_text().splitlines() if x.startswith('Test')]
for row in rows:
 kills=[id for id,rs in matrix.items() if row in rs];unique=[id for id in kills if len(matrix[id])==1];subs=[];seconds=None
 if unique:verdict='slow-worthy' if med[row]>60 else 'sacred'
 elif not kills:verdict='untrue'
 else:
  candidates=[q for q in allrows if q!=row and all(q in matrix[id] for id in kills)]
  if candidates:
   q=min(candidates,key=lambda q:med.get(q,float('inf')))
   if q not in med:
    vals=[]
    for i in range(3):
     path=E/f'timing-{q}-{i}.log'
     with path.open('w') as out:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+q+'$'],cwd=R,stdout=out,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u042/cache/subsumer'))
     vals.append(next(e['Elapsed'] for e in reversed(events(path)) if not e.get('Test') and e.get('Action')=='pass'))
    med[q]=round(statistics.median(vals),3)
   verdict='subsumed';subs=[q];seconds=med[q]
  else:verdict='overlapping';subs=sorted({q for id in kills for q in matrix[id] if q!=row})
 last=kills[-1] if kills else None;line=failure(logs[last,row],row) if last else None;own=[p[0] for p in json.load(open(E/'probes.json')) if row in p[4]];pk=[id for id in own if row in pmat[id]]
 file='predicates_overload_test.go' if row in rows[:2] else 'predicates_test.go' if row in rows[9:11] else 'primitive_slots_test.go' if row==rows[-1] else 'predicates_proof_test.go'
 oracle='Handwritten proof/admission/IR expectations; no external authority checked.'
 if row=='TestPredicateOverloadRuntime':oracle='Node runs original fixtures and validates handwritten stdout; native and backend JS match those validated values. Checked exit 70, diagnostic text and counters are self-written contracts.'
 if row=='TestPredicateBodyProof':oracle+=' TypeScript 6.0.3 supplies positive input bodies, not expected Adamic proof results.'
 if row=='TestIndirectPredicateOverloadIsPending':oracle+=' Checks diagnostic substring, not diagnostic type.'
 if not kills:oracle+=' Only the assignment-flow mutation reached the direction checker chain; no production kill was demonstrated in this plan.'
 cmd=('ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/'+last+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '+('^'+row+'$' if last in bounds else '.')+' > '+logs[last,row].name+' 2>&1') if last else None
 r=dict(test=row,package='internal/lower',file='internal/lower/'+file,seconds=med[row],oracle=oracle,oracle_kind=['external-run','self'] if row==rows[0] else 'self',kills=kills,unique_kills=unique,last_proven_fail=last+': '+line if line else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=20,probe_kills=pk,subsumer_seconds=seconds,vacuous=(not pk) if own else None,bounded=bool(bounds),matrix_rows=rows if bounds else [],bounded_mutants=bounds,evidence=cmd+'; '+line if cmd and line else 'No production mutant failure observed; all twenty columns and own-entry probe logs saved.')
 if vsub.get(row):r['vacuous_subcases']=vsub[row]
 if verdict=='subsumed':r['subsumption_mutants']=len(kills)
 result.append(r)
(E/'rows.json').write_text(json.dumps(result,indent=2));(E/'medians.json').write_text(json.dumps(med,indent=2))
with (E/'apply-check.log').open('w') as log:
 for p in plans:assert subprocess.run(['git','apply','--check','--cached',str(E/'diffs'/(p['id']+'.diff'))],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode==0
start=time.monotonic()
with (E/'clean-build.log').open('w') as log:rc=subprocess.run(['timeout','90','go','test','-c','-o','/tmp/u042-clean.test','./internal/lower/'],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
(E/'clean-build-time.json').write_text(json.dumps({'seconds':time.monotonic()-start,'exit':rc}))
(E/'report.done').write_text('done')
