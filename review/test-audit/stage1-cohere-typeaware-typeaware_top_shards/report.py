import json,pathlib,statistics,re
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-typeaware-typeaware_top_shards');groups=json.loads((p/'groups.json').read_text());plan=json.loads((p/'plan.json').read_text());runs=json.loads((p/'runs.json').read_text());timings=json.loads((p/'timings.json').read_text());byid={x['id']:x for x in plan};matrix=[];builds=[]
for r in runs:
 es=[]
 for line in (p/r['log']).read_text().splitlines():
  try:es.append(json.loads(line))
  except ValueError:pass
 failures=[e for e in es if e.get('Action')=='fail' and e.get('Test')];passes=[e.get('Test') for e in es if e.get('Action')=='pass' and e.get('Test')];m=byid[r['id']]
 errors=[]
 for e in es:
  if e.get('OutputType')=='error':
   o=e['Output'].strip();site=re.search(r'([\w]+_test\.go):(\d+):',o)
   if site and m['file'].endswith(site[1]) and int(site[2])>m['line']:
    delta=m['old'].count('\n')-m['new'].count('\n')
    if delta:o=o.replace(site[0],site[1]+':'+str(int(site[2])+delta)+':')+' [origin line; log line '+site[2]+']'
   errors.append({'test':e.get('Test'), 'line':o})
 matrix.append(dict(id=r['id'],log=r['log'],kind=m['kind'],invalid_reason=r.get('invalid_reason'),rows=r['rows'],failed_tests=[e['Test'] for e in failures],passed_tests=passes,unknown_tests=[n for n in r['rows'] if n not in passes and n not in [e['Test'] for e in failures]],errors=errors,corpus=r.get('corpus'),command=r['command']))
 for e in es:
  o=e.get('Output','')
  if 'phase clang ' in o or ('build volume-guard-' in o and (' miss ' in o or ' uncached ' in o)):
   builds.append(dict(id=r['id'],log=r['log'],test=e.get('Test'),line=o.strip(),invalid=bool(r.get('invalid_reason'))))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'rebuild-times.json').write_text(json.dumps(builds,indent=2)+'\n')
valid=[x for x in matrix if not x['invalid_reason']]
def fails(group,id):return any(x['id']==id and any(t in groups[group] for t in x['failed_tests']) for x in valid)
def proof(group,id):
 xs=[x for x in valid if x['id']==id and any(t in groups[group] for t in x['failed_tests'])]
 if not xs:return None,None
 x=xs[-1];err=next((e['line'] for e in x['errors'] if e['test'] in groups[group]),'--- FAIL: '+next(t for t in x['failed_tests'] if t in groups[group]));return x,err
ordinary=['TestTypeAwareAgreementAndMutants family','TestVolumeConfigGuardAndMutant family','TestVolumeConfigGuardAndMutant_000'];kills={g:[m for m in ['M1','M2','M3','M4'] if fails(g,m)] for g in ordinary};rows=[]
roles={'TestTypeAwareTopPlanted family':('setup-check','S1'),'TestTypeAwareTopShardPlantedFailure':('witness','W1'),'TestTypeAwareAgreementAndMutants_Setup':('setup-check','S2'),'TestProduct_volume_guard family':('setup-check','S5'),'TestVolumeConfigGuardAndMutant_Setup':('setup-check','S4'),'TestVolumeConfigGuardAndMutantPlantedFailure':('witness','W2')}
for g,members in groups.items():
 seconds=statistics.median(x['seconds'] for x in timings if x['group']==g);ks=kills.get(g,[]);unique=[m for m in ks if sum(m in v for v in kills.values())==1];subs=[];subseconds=None
 if g in roles:
  intended,id=roles[g];verdict=intended if fails(g,id) else 'cannot-judge';pr=id if fails(g,id) else None
 else:
  if unique:verdict='slow-worthy' if seconds>60 else 'sacred'
  elif not ks:verdict='untrue'
  else:
   sup=[h for h in ordinary if h!=g and set(ks)<=set(kills[h])]
   if sup:
    h=min(sup,key=lambda h:statistics.median(x['seconds'] for x in timings if x['group']==h));subs=[h];subseconds=statistics.median(x['seconds'] for x in timings if x['group']==h);verdict='subsumed'
   else:verdict='overlapping';subs=[h for h in ordinary if h!=g and set(ks)&set(kills[h])]
  pr=ks[-1] if ks else None
 x,err=proof(g,pr) if pr else (None,None)
 filename='typeaware_top_shards_test.go' if g.startswith('TestTypeAware') else 'volume_config_guard_self_prepare_test.go'
 if g.endswith('_Setup') and g.startswith('TestTypeAware'):filename='setup_deadline_typeaware_test.go'
 if g.startswith('TestProduct'):filename='volume_config_guard_products_test.go'
 semantic=g in ordinary;oracle=('Unchanged Go cohere full finding-byte comparison, sanitizer stderr, plus self-written refusal text and census assertions. Benchmarks compare counts only.' if g.startswith('TestTypeAwareAgreement') and semantic else 'Unchanged Go cohere full finding bytes and self-written compiler-option refusal code 70 plus exact stderr.' if semantic else 'Self-written shard construction, expected child mismatch signature or successful product preparation.')
 kind=['external-run','self'] if semantic or g=='TestVolumeConfigGuardAndMutantPlantedFailure' else 'self'
 probes=[id for id in (['P1'] if g==ordinary[0] else ['P2'] if semantic else []) if fails(g,id)]
 row=dict(test=g,package='stage1/cohere/typeaware',file='stage1/cohere/typeaware/'+filename,members=members,seconds=seconds,oracle=oracle,oracle_kind=kind,kills=ks,unique_kills=unique,last_proven_fail=(pr+': '+err) if err else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=4,probe_kills=probes,subsumer_seconds=subseconds,vacuous=False if probes else None,bounded=True,matrix_rows=sorted(set(t for z in valid for t in z['rows'])),evidence=(' '.join(x['command'])+' > '+x['log']+' 2>&1; '+err) if x else 'No valid deciding run yet',setup_kills=[roles[g][1]] if g in roles and roles[g][0]=='setup-check' and fails(g,roles[g][1]) else [],witness_kills=[roles[g][1]] if g in roles and roles[g][0]=='witness' and fails(g,roles[g][1]) else [])
 if g in ordinary:row['subsumption_basis_mutants']=len(ks)
 rows.append(row)
(p/'results.json').write_text(json.dumps(rows,indent=2)+'\n')
print([(r['test'],r['verdict'],r['kills'],r['last_proven_fail']) for r in rows])
