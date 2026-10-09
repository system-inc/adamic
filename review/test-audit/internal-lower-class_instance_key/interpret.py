import pathlib,json,re,csv,statistics
p=pathlib.Path(__file__).parent
assigned=json.loads((p/'rows.json').read_text()); menu=json.loads((p/'menu.json').read_text()); tests=[r for r in (p/'list.log').read_text().splitlines() if r.startswith('Test')]; families=json.loads((p/'families.json').read_text()); membermap={r:'TestOverrideParameterRefusal family' for rs in families.values() for r in rs}; package_rows=sorted(set(membermap.get(r,r) for r in tests))
def ev(name):
 f=p/(name+'.log')
 if not f.exists():return []
 result=[]
 for line in f.read_text().splitlines():
  if line.startswith('{'):
   try:result.append(json.loads(line))
   except json.JSONDecodeError:pass
 return result
matrix={}; raw={}; bounded={}; failedlines={}; logs={}
for m in menu:
 mid=m['id']
 if m['kind']=='probe':continue
 if not (p/(mid+'.log')).exists():continue
 es=ev(mid); state={r:'unknown' for r in tests}; bounds=False
 for e in es:
  if e.get('Test') in state and e['Action'] in ['pass','fail','skip']:state[e['Test']]=e['Action']
  if e['Action']=='fail' and e.get('Test','').split('/')[0] in state:state[e['Test'].split('/')[0]]='fail'
 for r in assigned:
  name=mid+'-'+r
  if (p/(name+'.log')).exists():
   bounds=any(e['Action']=='output' and e.get('Output','').startswith('panic:') for e in es)
   for e in ev(name):
    if e.get('Test')==r and e['Action'] in ['pass','fail','skip']:state[r]=e['Action']
    if e['Action']=='fail' and e.get('Test','').split('/')[0]==r:state[r]='fail'
 name=mid+'-bounded'
 if (p/(name+'.log')).exists():
  bounds=True
  for e in ev(name):
   if e.get('Test') in state and e['Action'] in ['pass','fail','skip']:state[e['Test']]=e['Action']
 raw[mid]=state
 grouped={}
 for r in package_rows:
  members=[t for t in tests if membermap.get(t,t)==r]; values=[state[t] for t in members]
  grouped[r]='fail' if 'fail' in values else 'unknown' if 'unknown' in values else 'skip' if all(v=='skip' for v in values) else 'pass'
 matrix[mid]=grouped; bounded[mid]=bounds
 for r in assigned:
  name=mid+'-'+r if (p/(mid+'-'+r+'.log')).exists() else mid+'-bounded' if (p/(mid+'-bounded.log')).exists() else mid
  lines=[e['Output'].strip() for e in ev(name) if e['Action']=='output' and e.get('Test','').split('/')[0]==r and (e.get('OutputType','').startswith('error') or 'panic:' in e.get('Output',''))]
  if grouped[r]=='fail' and not any(re.search(r'_test.go:\d+:',l) for l in lines):
   lines += [e['Output'].strip() for e in ev(name) if e['Action']=='output' and e.get('Test','').split('/')[0]==r and '_test.go:' in e.get('Output','')]
  if grouped[r]=='fail':failedlines[(r,mid)]=next((l for l in lines if re.search(r'_test.go:\d+:',l)),next(iter(lines),'row fail event; see '+name+'.log'));logs[(r,mid)]=name
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n'); (p/'matrix-raw.json').write_text(json.dumps(raw,indent=2)+'\n')
with (p/'matrix.csv').open('w') as f:
 w=csv.writer(f);w.writerow(['row',*matrix]);w.writerows([r,*[matrix[m][r] for m in matrix]] for r in package_rows)
base={}
for e in ev('enabled-baseline'):
 if e.get('Test') in tests and e['Action']=='pass':base[e['Test']]=e.get('Elapsed',0)
locations={}
for f in pathlib.Path('internal/lower').glob('*_test.go'):
 s=f.read_text()
 for m in re.finditer(r'^func (Test\w+)\(',s,re.M):locations[m[1]]=str(f)+':'+str(s[:m.start()].count('\n')+1)
def median(r):
 samples=[]
 for i in range(1,4):
  f=p/(r+'-'+str(i)+'.log')
  if not f.exists():return None
  m=re.search(r'\t([0-9.]+)s',f.read_text())
  if not m:return None
  samples.append(float(m[1]))
 return statistics.median(samples)
oracles={
 assigned[0]:'Self-written class-instantiation counts for checker-equivalent and distinct arguments. Counts do not verify class contents or native behavior.',
 assigned[1]:'Self-written iterator-receiver refusal label plus accepted unchanged-protocol controls. Does not execute iterator output.',
 assigned[2]:'Self-written acceptance expectation for an explicit string-key copy. Checks only absence of an error.',
 assigned[3]:'Self-written acceptance expectation for instance delegation of private storage. Checks only absence of an error.',
 assigned[4]:'Self-written shape acceptance and NotYet/Refused stop expectations. Unsupported subcases accept either stop type and can pass on a different later refusal.',
 assigned[5]:'Self-written NotYet expectation for branded primitive results. Does not check the reason or representation and can pass on a different later NotYet.',
 assigned[6]:'Self-written direct signature-proof expectation known=false for distinct null and undefined. Does not require a positive control.',
 assigned[7]:'Self-written direct signature-proof expectation known=false for an index signature. Does not require a positive control.',
 assigned[8]:'Self-written counts of nonempty readiness tags on IR reads/properties. Does not verify tag text or runtime readiness behavior.',
 assigned[9]:'Self-written acceptance expectation for sound initialized neighbors and prose mentioning ts-ignore. Checks only absence of an error.',
 assigned[10]:'Self-written IR element kinds and presence of an empty literal. Union case permits Number or String, rather than one exact layout.',
 assigned[11]:'Self-written exact multiset of Number and Object generic empty-return IR element kinds.',
 assigned[12]:'Self-written allocation-proof/Checked bits and exactly one entries call in IR. Does not execute enumeration output.',
 assigned[13]:'Self-written NotYet type plus specific reason substrings for three record boundaries.'}
results=[]; extra=set()
for r in assigned:
 kills=[m for m in matrix if m!='M10' and matrix[m][r]=='fail']; unique=[m for m in kills if sum(v=='fail' for v in matrix[m].values())==1]
 candidates=[s for s in package_rows if s!=r and kills and all(matrix[m][s]=='fail' for m in kills)]
 if unique: verdict='sacred'; subs=[]
 elif not kills: verdict='cannot-judge' if r==assigned[5] else 'untrue';subs=[]
 elif candidates:
  # Prefer already measured candidates; among these choose the smallest observed median.
  measured=[s for s in candidates if median(s) is not None]
  mutual=[s for s in candidates if s in assigned and {m for m in matrix if matrix[m][s]=='fail'}==set(kills)]
  subs=[min(mutual,key=median) if mutual else min(measured,key=median) if measured else min(candidates,key=lambda s:base.get(s,999))];verdict='subsumed'
  if median(subs[0]) is None:extra.add(subs[0])
 else:
  verdict='overlapping';subs=[];uncovered=set(kills)
  while uncovered:
   choices=[s for s in package_rows if s!=r and s not in subs]; pick=max(choices,key=lambda s:len([m for m in uncovered if matrix[m][s]=='fail']))
   covered={m for m in uncovered if matrix[m][pick]=='fail'}
   if not covered:break
   subs.append(pick);uncovered-=covered
 probe='P2' if r in assigned[6:8] else 'P1'; pe=ev(probe+'-'+r); status=next((e['Action'] for e in reversed(pe) if e.get('Test')==r and e['Action'] in ['pass','fail']),None)
 if any(e['Action']=='fail' and e.get('Test','').split('/')[0]==r for e in pe):status='fail'
 vac=None if status is None else status=='pass'
 last=kills[-1] if kills else None
 needs_bound=any(bounded[m] for m in kills)
 object=dict(test=r,package='internal/lower',file=locations[r],seconds=median(r),oracle=oracles[r],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=(last+' '+failedlines[(r,last)]) if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=len([m for m in matrix if m!='M10']),probe_kills=[probe] if status=='fail' else [],subsumer_seconds=median(subs[0]) if verdict=='subsumed' else None,vacuous=vac,bounded=needs_bound,evidence=(f'ADAMIC_MUTANT={last} ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/{last} timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '+('^'+r+'$' if logs.get((r,last),'').startswith(str(last)+'-Test') else '.')+'; '+failedlines[(r,last)]) if last else 'No production mutant failed this row in the observed matrix; see matrix.csv.')
 if r==assigned[5]:
  object['supplemental_kills']=['M10']; object['last_proven_fail']='M10 (supplemental): panic: runtime error: invalid memory address or nil pointer dereference; stack clock_generic_returns_t_01_test.go:66'
  object['evidence']='ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^'+r+'$; panic: runtime error: invalid memory address or nil pointer dereference; stack clock_generic_returns_t_01_test.go:66'
  object['cannot_judge_reason']='The only meaningful challenge to the branded type-argument guard removed condition terms, outside the fixed menu. It is supplemental. No admissible mutation meaningfully challenged that guard; no worthy verdict rests on M10.'
  object['bounded']=True;needs_bound=True
 if needs_bound:object['matrix_rows']=assigned
 if r==assigned[4]:object['vacuous_subcases']=[e['Test'].split('/',1)[1] for e in pe if e['Action']=='pass' and e.get('Test','').startswith(r+'/')]
 object['subsumption_mutants']=len(kills) if verdict=='subsumed' else None
 results.append(object)
(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');(p/'extra-timings-needed.json').write_text(json.dumps(sorted(extra),indent=2)+'\n')
print(json.dumps([{'test':r['test'],'kills':r['kills'],'unique':r['unique_kills'],'verdict':r['verdict'],'subsumed_by':r['subsumed_by'],'vacuous':r['vacuous']} for r in results],indent=2))
