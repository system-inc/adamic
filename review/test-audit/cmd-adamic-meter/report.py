import pathlib,json,re,statistics,csv
p=pathlib.Path('/tmp/u009');names=json.loads((p/'names.json').read_text());ms=json.loads((p/'mutants.json').read_text())
def events(f):
 a=[]
 if not f.exists():return a
 for l in f.read_text().splitlines():
  try:a.append(json.loads(l))
  except:pass
 return a
matrix={};evidence={};subs={};timedout={}
for m in ms:
 id=m['id'];ev=events(p/(id+'-matrix.log'));results={};err={}
 for n in names:
  solo=p/(id+'-'+n+'.log');a=events(solo) if solo.exists() else ev
  state=next((x['Action'] for x in reversed(a) if x.get('Test')==n and x['Action'] in ['pass','fail','skip']),None)
  timeout=any('test timed out' in x.get('Output','') for x in a)
  if timeout and state is None:state='over-budget'
  if state is None:
   # A panic on a row is an observed failure, not an assertion catch for unrelated rows.
   panic=any('panic:' in x.get('Output','') for x in a)
   if solo.exists() and panic:state='fail'
  results[n]=state
  outputs=[x.get('Output','').strip() for x in a if x.get('Test','').split('/')[0]==n]
  candidates=[x for x in outputs if re.search(r'_test.go:\d+:',x)]
  if candidates:err[n]=candidates[0]
  elif state=='fail':err[n]=next((x.get('Output','').strip() for x in a if 'panic:' in x.get('Output','')),'observed failure')
 matrix[id]=results;evidence[id]=err
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
seconds={}
for n in names:
 samples=[]
 for i in range(3):
  txt=(p/(n+'-'+str(i)+'.log')).read_text();samples.append(float(re.search(r'\bok\s+\S+\s+([\d.]+)s',txt).group(1)))
 seconds[n]=statistics.median(samples)
production=[m['id'] for m in ms if m['id'].startswith('M')];probes=[m['id'] for m in ms if m['id'].startswith('E')]
kills={n:[id for id in production if matrix[id][n]=='fail'] for n in names}
own={n:['E_ADAPT'] for n in names};own[names[0]]=['E_MEASURE'];own[names[1]]=['E_MEASURE'];own[names[2]]=['E_RUN'];own[names[3]]=['E_NORMALIZE'];own[names[11]]=['E_ADAPT','E_RETURN'];own[names[12]]=['E_RETURN'];own[names[13]]=['E_RETURN']
rows=[]
for n in names:
 k=kills[n];unique=[id for id in k if sum(x=='fail' for x in matrix[id].values())==1];other=[x for x in names if x!=n and k and set(k)<=set(kills[x])];subsumer=min(other,key=lambda x:seconds[x]) if other else None
 verdict='sacred' if unique else 'untrue' if not k else 'subsumed' if subsumer else 'overlapping'
 f='main_test.go' if n in names[:4] else 'optional_test.go' if n in names[4:11] else 'returns_test.go'
 oracle='Self-written report counts and labels.';kind='self'
 if n==names[1]:oracle='Self-written mechanical-diagnostic and rewrite counts, plus disk equality. It checks counts, not which diagnostic was removed.'
 if n==names[2]:oracle='Self-written report counts after JSON decoding; reason content is not checked.'
 if n==names[3]:oracle='Self-written normalized string.'
 if n in names[4:]:
  oracle='Linked TypeScript checker acceptance/diagnostic results plus self-written rewrite counts and no-change expectations; diagnostic-code counts do not verify full diagnostic identity.';kind=['external-run','self']
 if n in [names[4],names[9],names[11]]:oracle+=' Node runs the source and compares stdout to a self-written literal.'
 if n==names[11]:oracle+=' Node runs the original source, not an adapted product; E_ADAPT and E_RETURN both pass.'
 if n in [names[12],names[13]]:oracle+=' Current inputs have no TS7030 diagnostics; return adaptation exits at its initial guard.'
 last=k[-1] if k else None
 command=('ADAMIC_MUTANT='+last+' timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-meter/ -run . > '+last+'-matrix.log 2>&1') if last else 'All 20 production matrices and any required isolated reruns; no observed production-mutant failure.'
 fail=(last+': '+evidence[last].get(n,'failure recorded in log')) if last else None
 row=dict(test=n,package='github.com/system-inc/adamic/cmd/adamic-meter',file='cmd/adamic-meter/'+f,seconds=seconds[n],oracle=oracle,oracle_kind=kind,kills=k,unique_kills=unique,last_proven_fail=fail,verdict=verdict,subsumed_by=subsumer if verdict=='subsumed' else None,mutants_in_matrix=20,probe_kills=[id for id in probes if matrix[id][n]=='fail'],subsumer_seconds=seconds[subsumer] if verdict=='subsumed' else None,vacuous=all(matrix[id][n]=='pass' for id in own[n]),bounded=matrix['M16'][n]=='over-budget',matrix_rows=names if matrix['M16'][n]=='over-budget' else [],evidence=command+('; '+evidence[last].get(n,'failure recorded in log') if last else ''),own_entry_probes=own[n],over_budget_mutants=[id for id in production if matrix[id][n]=='over-budget'],vacuous_subcases=[])
 rows.append(row)
(p/'report.json').write_text(json.dumps(rows,indent=2))
with open(p/'matrix.csv','w') as f:
 w=csv.writer(f);w.writerow(['test']+production+probes)
 for n in names:w.writerow([n]+[matrix[id][n] for id in production+probes])
print([(x['test'],x['verdict'],x['kills'],x['vacuous'],x['bounded']) for x in rows])
