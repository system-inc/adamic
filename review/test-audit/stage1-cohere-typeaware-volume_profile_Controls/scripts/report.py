import json,pathlib,re,statistics
r=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-typeaware-volume_profile_Controls');scope=json.loads((r/'scope.json').read_text());times=json.loads((r/'seconds.json').read_text());p='stage1/cohere/typeaware/';allnames=scope['requested']+scope['replacement'];members=lambda prefix:[n for n in allnames if n.startswith(prefix) and n!=prefix+'Lower' and n!=prefix+'_Setup' and n not in ('TestVolumeTypeSymbolPlantedSurvivor','TestVolumeTypeSymbolCommandDeadline')]
def events(path):
 out=[]
 if not path.exists():return out
 for line in path.read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
specs=[('TestVolumeProfileControlsLower','lower','volume_profile_Controls_test.go','self','build recipe checks construction','K5','PsetupLower'),('TestVolumeProfileControls family','controls','volume_profile_Controls_test.go','external-run','unchanged Go cohere production rules; exact finding bytes and exits',None,'Pport'),('TestShadowIndexMissingBindingUnion','missing-union','volume_profile_Controls_test.go','self','case enumeration count and uniqueness; does not run missing-binding semantics','K2',None),('TestVolumeProfileCorpora_Setup','corpus-setup','split_volume_products_test.go','self','build recipe product availability; return value discarded','K6','PsetupCorpus'),('TestVolumeProfileCorpora family','corpus-bounded','volume_profile_Corpora_test.go','external-run','unchanged Go cohere production rules; exact finding bytes plain and ASan',None,'Pport'),('TestVolumeProfileMutants family','mutants','volume_profile_Mutants_test.go','external-run','Go cohere byte disagreement must catch built-in port mutants','W2',None),('TestVolumeProfileOverlays family','overlays','volume_profile_Overlays_test.go',['external-run','self'],'raw compiler AST identity checker run through Go; requires nonzero and specific diagnostic','W3',None),('TestVolumeProfilePartition','partition','volume_profile_partition_test.go','self','expected case ownership and shard wrapper names compared with parsed Go source','K3',None),('TestVolumeAgreementAndMutants family','agreement-bounded','volume_agreement_shards_test.go','external-run','Go cohere exact finding bytes; mixed bridge mutation witnesses outside bounded semantic leaves',None,'Pport'),('TestVolumeConfigGuardAndMutant','config-setup','volume_config_guard_self_prepare_test.go','self','preparation product repository nonempty; migrated row performs no config guard comparisons','K7','PsetupConfig'),('TestVolumeTypeSymbol family','symbol','volume_type_symbol_test.go','external-run','Go cohere exact byte disagreement on built-in checker mutant','W1',None),('TestVolumeTypeSymbolPlantedSurvivor','symbol-planted','volume_type_symbol_test.go','self','synthetic equal byte arrays must be reported as surviving mutant','W4',None),('TestVolumeTypeSymbolCommandDeadline','deadline','volume_type_symbol_test.go','self','100ms command cancellation, nonzero exit and elapsed under3s','K4',None)]
rows=[];matrixnames=set();kills={};evidence={};statuses={}
for log in r.glob('logs/M*-*.log'):
 mid=log.name.split('-')[0]
 if '-build' in log.name:continue
 for e in events(log):
  n=e.get('Test','');top=n.split('/')[0]
  if n:matrixnames.add(top)
  if e.get('Action') in ('pass','fail') and n==top:statuses.setdefault(mid,{})[top]=e['Action']
  if e.get('Action')=='fail' and n==top:kills.setdefault(top,set()).add(mid)
  if e.get('OutputType')=='error' or ('mismatch byte' in e.get('Output','')):evidence.setdefault((mid,top),(log.name,e.get('Output','').strip()))
for name,key,file,kind,oracle,experiment,probe in specs:
 if ' family' in name:
  prefix=name.split(' family')[0];ms=members(prefix)
  if prefix=='TestVolumeAgreementAndMutants':ms=scope['replacement']
 else:ms=[name]
 ks=sorted(set().union(*(kills.get(n,set()) for n in ms))) if experiment is None else []
 ek=[];last=None;proof=None
 if experiment:
  for e in events(r/'logs'/f'{experiment}.log'):
   if e.get('Test','').split('/')[0] in ms and (e.get('OutputType')=='error' or 'FAIL:' in e.get('Output','')):ek.append(e.get('Output','').strip())
  if ek:proof=f"{experiment}: {ek[0]}";last=proof
 for mid in ks:
  for n in ms:
   if (mid,n) in evidence:log,line=evidence[(mid,n)];proof=f'{mid}: {line}';last=proof;break
 probes=[];vacuous=None
 if probe:
  ev=events(r/'logs'/f'{probe}.log');actions=[e.get('Action') for e in ev if e.get('Test') in ms and e.get('Action') in ('pass','fail')]
  if actions:
   vacuous='fail' not in actions
   if not vacuous:probes=[probe]
 verdict=('witness' if experiment and experiment.startswith('W') else 'setup-check') if last and experiment else 'cannot-judge'
 if ks:verdict='overlapping'
 row=dict(test=name,package='stage1/cohere/typeaware',file=p+file,seconds=times[key]['median'],oracle=oracle,oracle_kind=kind,kills=ks,unique_kills=[],last_proven_fail=last,verdict=verdict,subsumed_by=[],mutants_in_matrix=4 if not experiment else 0,probe_kills=probes,subsumer_seconds=None,vacuous=vacuous,bounded=True,matrix_rows=['TestVolumeProfileControls family','TestVolumeProfileCorpora family','TestVolumeAgreementAndMutants family'],matrix_members=sorted(matrixnames),members=ms,evidence=proof,timing_samples=times[key]['samples'],timing_bounded=key in ('corpus-bounded','agreement-bounded'))
 if not experiment and not ks:row['verdict']='untrue' if all(any(statuses.get(mid,{}).get(n)=='pass' for n in ms if not n.endswith('Union')) for mid in ('M1','M2','M3','M4')) else 'cannot-judge'
 rows.append(row)
semantic=[x for x in rows if x['kills']]
for x in semantic:
 unique=[mid for mid in x['kills'] if sum(mid in z['kills'] for z in semantic)==1];x['unique_kills']=unique
 if unique:x['verdict']='sacred';continue
 subs=[z for z in semantic if z is not x and set(x['kills'])<=set(z['kills'])]
 if subs:
  z=min(subs,key=lambda z:z['seconds']);x.update(verdict='subsumed',subsumed_by=[z['test']],subsumer_seconds=z['seconds'])
 else:x['subsumed_by']=[z['test'] for z in semantic if z is not x and set(x['kills'])&set(z['kills'])]
for x in rows:
 x['unknown_cells']={mid:[n for n in x['members'] if n in matrixnames and n not in statuses.get(mid,{})] for mid in ('M1','M2','M3','M4')} if x['mutants_in_matrix'] else {}
 if x['evidence']:
  experiment=next(z[5] for z in specs if z[0]==x['test'])
  recs=json.loads((r/('experiment-runs.json' if experiment else 'matrix-runs.json')).read_text())
  ident=experiment or x['last_proven_fail'].split(':')[0]
  rec=next(z for z in recs if z['id']==ident and (experiment or z['phase'] in ('matrix','bounded')))
  x['evidence']='timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '+repr(rec['pattern'])+'; '+x['evidence']
(r/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(r/'matrix.json').write_text(json.dumps(statuses,indent=2)+'\n')
(r/'final-array.json').write_text(json.dumps([{k:v for k,v in x.items() if k not in ('members','matrix_members','timing_samples','unknown_cells')} for x in rows],indent=2)+'\n')
print([(x['test'],x['verdict'],x['kills'],x['vacuous']) for x in rows])
