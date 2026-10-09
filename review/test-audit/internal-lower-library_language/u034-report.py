import pathlib,json,re,subprocess,statistics
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/internal-lower-library_language'
rows=['TestLibraryLanguageBoundaries','TestLibraryMapSetGapsStayRefused','TestLibraryMapSetIteratorCopyTypesRefused','TestLibraryMethodValues','TestLibraryMethodValueBoundaries','TestLibraryMethodValueSafety','TestNodeBufferRefusals'];allrows=[l for l in (p/'u034-list.log').read_text().splitlines() if l.startswith('Test')]
def events(f):
 out=[]
 for l in f.read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
runs=json.loads((p/'runs.json').read_text());menu=json.loads((p/'menu.json').read_text());timings=json.loads((p/'timings.json').read_text());matrix={}
for run in runs:
 id=run['id'];ev=events(p/(id+'.log'));fail={e['Test'].split('/')[0] for e in ev if e.get('Action')=='fail' and e.get('Test')};observed={e['Test'] for e in ev if e.get('Action') in ['pass','fail','skip'] and e.get('Test') and '/' not in e['Test']}
 if run['panic'] or run['timeout']:
  for row in rows:
   f=p/(id+'-'+row+'.log');individual=events(f);observed.add(row)
   if any(e.get('Action')=='fail' for e in individual):fail.add(row)
 matrix[id]=dict(kills=sorted(fail),observed_rows=sorted(observed),unknown_rows=sorted(set(allrows)-observed),bounded=run['panic'] or run['timeout'],individual_reruns=rows if run['panic'] or run['timeout'] else [],command='ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR=/tmp/u034/cache/'+id+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '+run['pattern']+' > '+id+'.log 2>&1')
(p/'matrix.json').write_text(json.dumps(dict(package_rows=allrows,scoped_rows=rows,mutants=matrix,excluded='invalid-instrumentation/*'),indent=2))
prod=[m['id'] for m in menu if m['id'].startswith('M')];kills={r:[id for id in prod if r in matrix[id]['kills']] for r in allrows}
oracles=["Handwritten refusal substrings for Adamic language boundaries. An unrelated error containing the substring can pass.","Handwritten NotYet identity and refusal substrings for unsupported protocols and representations.","Handwritten NotYet identity and inherited-next refusal text. JavaScript property ownership explains the boundary; expected diagnostic is Adamic-owned.","Acceptance only; no returned IR or emitted behavior is inspected. Passes Lower's empty-answer probe.","Handwritten refusal substrings for method-value boundaries; unrelated refusal with matching text can pass.","Any nonnil error is rejection; unrelated refusals can mask unsafe admission. Production M16 actually admits an unsafe case and makes this row fail.","Handwritten refusal substrings for Node Buffer/Hash census. Does not run Node or inspect the rejected operation's behavior."]
result=[]
for i,row in enumerate(rows):
 k=kills[row];unique=[id for id in k if matrix[id]['kills']==[row] and not matrix[id]['bounded']]
 candidates=[r for r in allrows if r!=row and k and set(k)<=set(kills[r])]
 # Scoped candidate median is actually measured. Outside candidates would need their own timing runs.
 candidates=[r for r in candidates if r in timings]
 if unique:v='sacred';sub=[];secs=None
 elif candidates:v='subsumed';sub=[min(candidates,key=lambda r:timings[r]['median'])];secs=timings[sub[0]]['median']
 elif k:
  v='overlapping';sub=[];remaining=set(k)
  while remaining:
   other=max([r for r in allrows if r!=row],key=lambda r:len(remaining&set(kills[r])))
   covered=remaining&set(kills[other]);assert covered
   sub.append(other);remaining-=covered
  secs=None
 else:v='untrue';sub=[];secs=None
 chosen=unique[-1] if unique else k[-1] if k else None
 assertion='';fail=''
 if chosen:
  ev=events(p/(chosen+'.log'))
  if matrix[chosen]['bounded']:ev+=events(p/(chosen+'-'+row+'.log'))
  assertion=next((e['Output'].strip() for e in ev if e.get('Test','').split('/')[0]==row and re.search(r'_test.go:\d+:',e.get('Output',''))),'')
  fail=next((e['Output'].strip() for e in ev if e.get('Test')==row and e.get('Output','').startswith('--- FAIL:')),'')
  if not fail:fail=next((e['Output'].strip() for e in ev if 'panic:' in e.get('Output','')),'')
 bounded=any(matrix[id]['bounded'] for id in k)
 file='internal/lower/'+('library_language_test.go' if i==0 else 'library_map_set_test.go' if i<3 else 'library_method_values_test.go' if i<6 else 'library_node_buffer_test.go')
 result.append(dict(test=row,package='internal/lower',file=file,seconds=timings[row]['median'],oracle=oracles[i],oracle_kind='self',kills=k,unique_kills=unique,last_proven_fail=(chosen+' '+fail) if chosen else None,verdict=v,subsumed_by=sub,mutants_in_matrix=20,probe_kills=['P01'] if row in matrix['P01']['kills'] else [],subsumer_seconds=secs,vacuous=row not in matrix['P01']['kills'],bounded=bounded,matrix_rows=rows if bounded else allrows,evidence=(matrix[chosen]['command']+'; '+assertion) if chosen else 'No observed production-mutant failure.'))
(p/'results.json').write_text(json.dumps(result,indent=2))
for m in menu:
 rc=subprocess.run(['git','apply','--check',str(p/(m['id']+'.diff'))],cwd=root,capture_output=True);assert rc.returncode==0,(m['id'],rc.stderr)
summary=dict(starting_commit='e2492670b06a4dce1deafe837158ce0366cf2bc4',nproc=5,baseline_seconds=33.730,inactive_switch_seconds=28.049,verdicts={v:sum(x['verdict']==v for x in result) for v in ['sacred','subsumed','overlapping','untrue']},survivors=[id for id in prod if not matrix[id]['kills']],vacuous=[x['test'] for x in result if x['vacuous']],matrix_wall_seconds=sum(x['wall_seconds'] for x in runs),timing_binary_seconds=sum(sum(x['runs']) for x in timings.values()),vet_seconds=sum(x['seconds'] for x in json.loads((p/'validation.json').read_text())))
(p/'summary.json').write_text(json.dumps(summary,indent=2))
print(json.dumps(summary,indent=2))
for x in result:print(x['test'],x['kills'],x['verdict'],x['subsumed_by'])
