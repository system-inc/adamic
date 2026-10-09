import pathlib,json,re,statistics
p=pathlib.Path(__file__).parent; rows=json.loads((p/'rows.json').read_text())
def events(name):
 out=[]
 for s in (p/name).read_text().splitlines():
  try:out.append(json.loads(s))
  except:pass
 return out
matrix={}
for mid in [f'M{i:02}' for i in range(1,21)]+['P01','P02']:
 es=events(mid+'.log');matrix[mid]={'failed_tests':[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']], 'panic':any('panic:' in e.get('Output','') for e in es),'completed':any(e.get('Action') in ['pass','fail'] and not e.get('Test') for e in es)}
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
seconds={};trials={}
for row in rows:
 vals=[]
 for i in range(3):
  s=(p/(row+f'-{i}.log')).read_text();m=re.search(r'\t([0-9.]+)s',s);assert m and '\nFAIL' not in s,(row,i);vals.append(float(m[1]))
 seconds[row]=statistics.median(vals);trials[row]=vals
(p/'seconds.json').write_text(json.dumps(trials,indent=2))
allrows=sorted(set(rows+sum([v['failed_tests'] for v in matrix.values()],[])))
kills={r:[m for m in matrix if m.startswith('M') and r in matrix[m]['failed_tests']] for r in allrows}
for row in allrows:
 if row in seconds:continue
 vals=[]
 for i in range(3):
  f=p/(row+f'-{i}.log')
  if f.exists():
   m=re.search(r'\t([0-9.]+)s',f.read_text())
   if m:vals.append(float(m[1]))
 if len(vals)==3:seconds[row]=statistics.median(vals)
checks=['accepts fixed stat options and void returns; only checks err == nil','mutable widening refusal contains invariant-mutable','option-effects refusal contains evaluated expressions','void-value refusal contains fs void calls used as values','NotYet type and member substring only; M09 demonstrates rejection reason can be wrong','namespace fs import accepted; only checks err == nil','Refused type plus no-optional-widening substring','Refused type and isFile member substring','Buffer read/write borrow accepted; only checks err == nil','fixed rm/mkdtemp options accepted; only checks err == nil','NotYet type and overload member substring','NotYet type and qualified declaration name substring','type-only imports and user names accepted; only checks err == nil','handwritten two distinct member names from AST symbol owners','Refused type plus no-optional-widening substring; fixed mutant plan did not isolate provenRelation optional-field branch']
results=[]
for idx,row in enumerate(rows):
 ks=kills[row];unique=[m for m in ks if matrix[m]['failed_tests']==[row]];subs=[];subs_sec=None
 if unique:verdict='sacred'
 elif not ks:verdict='untrue'
 else:
  candidates=[r for r in allrows if r!=row and set(ks)<=set(kills[r])]
  if candidates:
   measured=[r for r in candidates if r in seconds]
   if not measured:raise RuntimeError('Measure external subsumer for '+row+': '+str(candidates))
   subs=[min(measured,key=lambda r:seconds[r])];subs_sec=seconds[subs[0]];verdict='subsumed'
  else:
   remain=set(ks)
   while remain:
    r=max([r for r in allrows if r!=row],key=lambda r:(len(remain & set(kills[r])),r in rows))
    assert remain & set(kills[r]);subs.append(r);remain-=set(kills[r])
   verdict='overlapping'
 pid='P02' if row=='TestNodeLibraryDistinguishesReceiverOwners' else 'P01';pes=events(pid+'-'+row+'.log');probe_fail=any(e.get('Action')=='fail' and e.get('Test')==row for e in pes);assert any(e.get('Action') in ['pass','fail'] and e.get('Test')==row for e in pes),row
 def failure(es):
  return next((e['Output'].strip() for e in es if (e.get('Test')==row or e.get('Test','').startswith(row+'/')) and re.search(r'_test.go:\d+:',e.get('Output',''))),None)
 last=ks[-1] if ks else None;line=failure(events(last+'.log')) if last else failure(pes)
 cmd='timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .'
 if not last:cmd='ADAMIC_MUTANT='+pid+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^'+row+'$'
 else:cmd='ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/'+last+' '+cmd
 results.append(dict(test=row,package='internal/lower',file='internal/lower/'+('library_node_fs_file_test.go' if idx<11 else 'library_node_test.go'),seconds=seconds[row],oracle='Handwritten assertion: '+checks[idx],oracle_kind='self',kills=ks,unique_kills=unique,last_proven_fail=(last+': '+str(line)) if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=20,probe_kills=[pid] if probe_fail else [],subsumer_seconds=subs_sec,vacuous=not probe_fail,bounded=False,matrix_rows=[],evidence=cmd+' > '+(last or pid+'-'+row)+'.log 2>&1; '+str(line)))
(p/'results.json').write_text(json.dumps(results,indent=2))
print(json.dumps([(o['test'],o['verdict'],o['unique_kills'],o['vacuous']) for o in results],indent=2))
