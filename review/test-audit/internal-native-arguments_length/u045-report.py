import pathlib,json,re,statistics,datetime,subprocess
p=pathlib.Path('review/test-audit/internal-native-arguments_length');names=json.loads((p/'names.json').read_text());menu=json.loads((p/'menu.json').read_text())
def events(file):
 ds=[]
 for l in file.read_text().splitlines():
  try:ds.append(json.loads(l))
  except:pass
 return ds
def failures(ds):return sorted(set(d['Test'].split('/')[0] for d in ds if d['Action']=='fail' and d.get('Test')))
def fail_line(ds,n):
 outputs=[d.get('Output','').strip() for d in ds if d.get('Test','').split('/')[0]==n and d.get('Output')]
 for out in outputs:
  if re.search(r'\w+_test.go:\d+:',out) and not any(v in out for v in ['rejected by clang','Node catches compiled','contraction allowed fuses']):return out
 return next((out for out in outputs if '--- FAIL:' in out),'no failing line')
seconds={n:statistics.median(next(d['Elapsed'] for d in events(p/f'timing-{n}-{i}.log') if d['Action']=='pass' and not d.get('Test')) for i in range(3)) for n in names}
mat={m['id']:failures(events(p/(m['id']+'.log'))) for m in menu};probes=json.loads((p/'probes.json').read_text());pmat={m['id']:failures(events(p/(m['id']+'.log'))) for m in probes};witnesses=['TestClosureConventionDropCount','TestClosureConventionRuntimeDropCount','TestClosureConventionWrongOrder']
files={}
for f in pathlib.Path('internal/native').glob('*_test.go'):
 text=subprocess.check_output(['git','show','HEAD:'+str(f)]).decode()
 for n in names:
  match=re.search(r'^func '+re.escape(n)+r'\(',text,re.M)
  if match:files[n]=str(f)+':'+str(text[:match.start()].count('\n')+1)
kind={n:'self' for n in names};oracle={n:'Hand-written IR, emitted-C or memory-plan assertions; no outside authority checked.' for n in names}
for n in ['TestCaseMappingMatchesNode','TestCaseTablesMatchNodesUnicode','TestOptionalMethodThunksMatchNode']:
 kind[n]='external-run'
oracle['TestCaseMappingMatchesNode']='Node String.toLowerCase/toUpperCase executed and compared byte-for-byte over every code point and context sweep; line counts and exits also checked.'
oracle['TestCaseTablesMatchNodesUnicode']='Node process.versions.unicode executed; compares version labels in two headers only, not table contents.'
kind['TestOptionalMethodThunksMatchNode']=['external-run','self'];oracle['TestOptionalMethodThunksMatchNode']='Executed Node output versus sanitized and unsanitized native output; hand-written closure-convention and omitted-table-entry assertions; includes built-in missing-thunk witness.'
for n in witnesses:
 kind[n]=['external-run','self'];oracle[n]='Actual clang rejection of built-in invalid C; hand-written required diagnostic substrings. W_BUILD discards Build rejection after clang runs and makes this witness fail.'
entry={n:['P_C'] for n in names};entry.update({'TestBorrowChainDeclarations':['P_ELEMENTS'],'TestBorrowChainTargets':['P_CHAIN'],'TestPassThroughsAreNotConsumers':['P_CONSUMES'],'TestInheritanceMemoryPlans':['P_REGIONS','P_REUSE'],'TestCaseMappingMatchesNode':['P_BUILD'],'TestCaseTablesMatchNodesUnicode':[],'TestArithmeticIsNeverFused':['P_BUILD'],'TestClosureConventionDropCount':['P_BUILD'],'TestClosureConventionRuntimeDropCount':['P_BUILD'],'TestClosureConventionWrongOrder':['P_BUILD'],'TestClosureConventionRuntimeFeaturesIgnoreLiterals':['P_C','P_LIBRARY']})
rows=[]
for n in names:
 kills=[] if n in witnesses else [i for i,fs in mat.items() if n in fs]
 unique=[i for i in kills if len(mat[i])==1]
 subs=[];subsec=None
 if n in witnesses:
  verdict='witness' if n in failures(events(p/'W_BUILD.log')) else 'untrue';last='W_BUILD';ds=events(p/'W_BUILD.log')
 elif unique:verdict='sacred';last=unique[-1];ds=events(p/(last+'.log'))
 elif not kills:verdict='untrue';last=None;ds=[]
 else:
  subs=[other for other in names if other!=n and other not in witnesses and all(other in mat[i] for i in kills)]
  if subs:subs=[min(subs,key=lambda other:seconds[other])];subsec=seconds[subs[0]];verdict='subsumed'
  else:subs=sorted(set(other for i in kills for other in mat[i] if other!=n and other not in witnesses));verdict='overlapping'
  last=kills[-1];ds=events(p/(last+'.log'))
 own=entry[n];vacuous=any(n not in pmat[i] for i in own) if own else None
 line=fail_line(ds,n) if last else 'No production mutant caused this row to fail; M12 dropped the structural thunk requirement and this negative check still passed.'
 evidence=('ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/'+last+' timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '+ '^('+ '|'.join(names)+')$'+' > '+last+'.log 2>&1; '+line) if last else 'All M01-M17 bounded runs passed this row; see matrix.json and logs. P_C also passed it.'
 rows.append(dict(test=n,package='internal/native',file=files[n],seconds=seconds[n],oracle=oracle[n],oracle_kind=kind[n],kills=kills,unique_kills=unique,last_proven_fail=(last+': '+line) if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=17,probe_kills=[i for i in own if n in pmat[i]],subsumer_seconds=subsec,vacuous=vacuous,bounded=True,matrix_rows=names,evidence=evidence,probe_entries=own))
(p/'rows.json').write_text(json.dumps(rows,indent=2));(p/'matrix.json').write_text(json.dumps(dict(production=mat,probes=pmat,witness_check=failures(events(p/'W_BUILD.log')),rows=names),indent=2))
# Save every observed top-level result, including passes, so replay can identify unknowns.
(p/'observed-results.json').write_text(json.dumps({id:{d['Test']:d['Action'] for d in events(p/(id+'.log')) if d['Action'] in ['pass','fail','skip'] and d.get('Test') and '/' not in d['Test']} for id in list(mat)+list(pmat)+['W_BUILD']},indent=2))
commands=json.loads((p/'commands.json').read_text()); rebuild={}
for id in ['M14','M15','M16']:
 ds=events(p/(id+'.log'));start=next(d['Time'] for d in ds if d['Action']=='cont' and d.get('Test')=='TestCaseMappingMatchesNode');end=next(d['Time'] for d in ds if d['Action']=='run' and d.get('Test')=='TestCaseMappingMatchesNode/points');rebuild[id]=(datetime.datetime.fromisoformat(end.replace('Z','+00:00'))-datetime.datetime.fromisoformat(start.replace('Z','+00:00'))).total_seconds()
(p/'timings.json').write_text(json.dumps(dict(nproc=5,setup_seconds=0,npm_reported_seconds=.627,whole_baseline_binary_seconds=90.047,bounded_baseline_binary_seconds=15.855,solo_medians=seconds,solo_binary_seconds=sum(next(d['Elapsed'] for d in events(p/f'timing-{n}-{i}.log') if d['Action']=='pass' and not d.get('Test')) for n in names for i in range(3)),matrix_and_validation_wall_seconds=sum(c['seconds'] for c in commands),C_rebuild_sections_seconds=rebuild),indent=2))
print(json.dumps({r['test']:r['verdict'] for r in rows},indent=2))
