import pathlib,json,re,statistics
p=pathlib.Path(__file__).parent;rows=json.loads((p/'rows.json').read_text());status=json.loads((p/'matrix-status.json').read_text())
def events(name):
 es=[]
 for s in (p/name).read_text().splitlines():
  try:es.append(json.loads(s))
  except:pass
 return es
matrix={}
for mid,st in status.items():
 es=events(mid+'.log');assert any(e.get('Action') in ['pass','fail'] and not e.get('Test') for e in es),(mid,'incomplete')
 assert not any('No space left' in e.get('Output','') or "undeclared function 'getenv'" in e.get('Output','') for e in es),mid
 matrix[mid]={'rows_run':st['rows'],'failed_rows':[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']],'binary_seconds':next(e.get('Elapsed') for e in reversed(es) if e.get('Action') in ['pass','fail'] and not e.get('Test')),'command':st['command'],'wall_seconds':st['wall']}
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));seconds={};trials={}
for row in rows:
 vals=[]
 for i in range(3):
  text=(p/(row+f'-{i}.log')).read_text();assert not text.startswith('--- FAIL'),row;vals.append(float(re.search(r'\t([0-9.]+)s',text)[1]))
 trials[row]=vals;seconds[row]=statistics.median(vals)
(p/'seconds.json').write_text(json.dumps(trials,indent=2));kills={r:[m for m,v in matrix.items() if m.startswith('M') and r in v['failed_rows']] for r in rows};results=[]
oracles=['Handwritten metamorphic key changes and flag order checks.','Handwritten empty-program count report, run from the built executable.','Handwritten C fixture outputs 1, 2, 4 and changed archive identities.','Handwritten one-compilation, identical archives, output 42 and one cache entry.','Handwritten subprocess header, output 42, successful child checks and one cache entry; parent is not merely a helper.','V8 13.6.233.17 include/v8-primitive.h:126 64-bit kMaxLength formula checked: (1<<29)-24=536870888. Adamic panic prefix and exit 70 are self-written.','Handwritten borrowing, local ownership, lending and function/loop inventory.','Handwritten IR Borrowed flags and absence of named binding retain/release in emitted C; no Node execution.','Handwritten eligible-global argument predicate and minimum coverage count.','Handwritten iterator retain/release expectations, emitted definitions and coverage counts.','Handwritten borrow/lending/iterator expectations, global-call checks and 26-function inventory.','C harness asserts SameValueZero, insertion/lookups and probe bound 64; Go checks exit status only. M20 changes hash(1) while this corpus still passes.','Witness expects three built-in bad hashes to fail with exit 1 and specific diagnostic text, without sanitizer failures.']
for idx,row in enumerate(rows):
 witness=row=='TestMapHashProbeCatchesMutants';ks=kills[row];unique=[m for m in ks if matrix[m]['failed_rows']==[row]];subs=[];subsec=None
 if witness:verdict='witness'
 elif unique:verdict='sacred'
 elif not ks:verdict='untrue'
 else:
  candidates=[r for r in rows if r!=row and set(ks)<=set(kills[r])]
  if candidates:sub=min(candidates,key=seconds.get);subs=[sub];subsec=seconds[sub];verdict='subsumed'
  else:
   verdict='overlapping';left=set(ks)
   while left:
    sub=max([r for r in rows if r!=row],key=lambda r:len(left&set(kills[r])));assert left&set(kills[sub]);subs.append(sub);left-=set(kills[sub])
 probes=[m for m,v in matrix.items() if m.startswith('P') and row in v['rows_run']];pk=[m for m in probes if row in matrix[m]['failed_rows']];vacuous=None if witness or not probes else not pk
 mid='W01' if witness else ks[-1] if ks else pk[-1] if pk else None;es=events(mid+'.log') if mid else [];line=next((e['Output'].strip() for e in es if (e.get('Test')==row or e.get('Test','').startswith(row+'/')) and re.search(r'_test.go:\d+:',e.get('Output',''))),None)
 if witness:assert any(e.get('Action')=='fail' and e.get('Test')==row for e in es),'witness did not fail'
 ran=[m for m,v in matrix.items() if m.startswith('M') and row in v['rows_run']];mr=matrix[ran[0]]['rows_run'] if ran else [row]
 command=matrix[mid]['command'] if mid in matrix else 'timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestMapHashProbeCatchesMutants$'
 last=('W01: '+str(line)) if witness else (ks[-1]+': '+str(line)) if ks else None
 result=dict(test=row,package='internal/native',file='internal/native/'+('library_test.go' if idx<5 else 'limit_test.go' if idx==5 else 'loop_borrow_test.go' if idx<11 else 'map_hash_test.go'),seconds=seconds[row],oracle=oracles[idx],oracle_kind=['external-authority','self'] if idx==5 else 'self',kills=ks,unique_kills=unique,last_proven_fail=last,verdict=verdict,subsumed_by=subs,mutants_in_matrix=len(ran),probe_kills=pk,subsumer_seconds=subsec,vacuous=vacuous,bounded=True,matrix_rows=mr,evidence=('W01 checker diff applied; ' if witness else 'ADAMIC_MUTANT='+str(mid)+' ')+command+' > '+str(mid)+'.log 2>&1; '+str(line))
 if idx==5:result['vacuous_subcases']=['0 units','536870888 units']
 results.append(result)
(p/'results.json').write_text(json.dumps(results,indent=2));print(json.dumps([(o['test'],o['verdict'],o['kills'],o['unique_kills'],o['vacuous']) for o in results],indent=2))
