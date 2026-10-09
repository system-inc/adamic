import pathlib,json,statistics,subprocess,difflib,re
p=pathlib.Path('review/test-audit/internal-native-math'); plan=json.loads((p/'plan.json').read_text()); matrix=json.loads((p/'matrix.json').read_text()); family=json.loads((p/'family.json').read_text()); tests=['TestMathAndToFixedMatchJavaScript','TestMaybeNumbersPackIntoOneDouble','TestNodeBufferRuntimeWithoutDeclarations','TestNormalizeLongMeasurements','TestNormalizeRandomMatchesNode']; matrixrows=[t for t in tests if t!='TestNormalizeLongMeasurements']; members=['TestNormalizeMatchesNodePoints%02d'%i for i in range(17)]+['TestNormalizeMatchesNodeContexts']; fam='TestNormalizeMatchesNode family'
def events(file):
 rs=[]
 for l in file.read_text().splitlines():
  try:rs.append(json.loads(l))
  except:pass
 return rs
def fails(file):return sorted(set(e['Test'].split('/')[0] for e in events(file) if e.get('Action')=='fail' and e.get('Test')))
for r in matrix:
 if r['id'] in ['M11','M12','M13','P09'] and (p/('family-'+r['id']+'.log')).exists():
  r['family_failed_members']=fails(p/('family-'+r['id']+'.log'));r['family_failed']=bool(r['family_failed_members'])
(p/'combined-matrix.json').write_text(json.dumps(matrix,indent=2))
seconds={}
for t in tests:
 vals=[]
 for i in range(3):
  file=p/f'timing-{t}-{i}.log'
  if file.exists():vals +=[e['Elapsed'] for e in events(file) if e.get('Action')=='pass' and e.get('Test')==t]
 seconds[t]=statistics.median(vals) if len(vals)==3 else None
famvals=[events(p/f'family-family-control-{i}.log')[-1].get('Elapsed') for i in range(3)];famsec=statistics.median(famvals)
files=['math_test.go','maybe_test.go','node_buffer_test.go','normalize_benchmark_test.go','normalize_random_test.go'];oracles=['Node Math.round/sign/max/min, exponentiation, Number.toFixed; exact answers including negative zero','Self-written reserved and canonical NaN bits, presence flag and exact round-trip bits; Go math.IsNaN only classifies inputs','Node v24.19.0 Buffer and crypto.createHash; exact stdout in release and sanitizer builds, plus JavaScript backend agreement','Self-written expected exit-status schedule. Node measurements are recorded but outputs and times are never compared; any nonzero oversized-case exit is accepted, including unrelated failures','Node String.normalize in all four forms, exact WTF-8 hex for 100000 independently generated input strings and output line count']
probeids=[['P%02d'%i for i in range(1,7)],['P07','P08'],['P%02d'%i for i in range(10,16)],['P09'],['P09']]
rows=[]
for t,file,oracle,own in zip(tests,files,oracles,probeids):
 kills=[r['id'] for r in matrix if r['id'].startswith('M') and t in r['failures']];pk=[r['id'] for r in matrix if r['id'] in own and t in r['failures']]
 unique=[r['id'] for r in matrix if r['id'] in kills and r['failures']==[t] and not r.get('family_failed')]
 last=next((r for r in reversed(matrix) if r['id'] in kills),None);line=None
 if last:
  outputs=[e.get('Output','').strip() for e in events(p/(last['id']+'.log')) if e.get('Test')==t];line=next((o for o in outputs if re.search(r'(answers differ|values pack wrong|lines differ)$',o)),None) or next((o for o in outputs if o.startswith('--- FAIL:')),None)
 verdict='sacred' if unique else ('subsumed' if t.endswith('RandomMatchesNode') and kills and all(next(r for r in matrix if r['id']==k).get('family_failed') for k in kills) else 'cannot-judge')
 r=dict(test=t,package='internal/native',file='internal/native/'+file,seconds=seconds[t],oracle=oracle,oracle_kind='self' if 'MaybeNumbers' in t or 'LongMeasurements' in t else 'external-run',kills=kills,unique_kills=unique,last_proven_fail=(last['id']+' '+str(line)) if last else None,verdict=verdict,subsumed_by=[fam] if verdict=='subsumed' else [],mutants_in_matrix=len([r for r in matrix if r['id'].startswith('M')]),probe_kills=pk,subsumer_seconds=famsec if verdict=='subsumed' else None,vacuous=False if pk else None,bounded=True,matrix_rows=matrixrows+([fam] if t.endswith('RandomMatchesNode') else []),evidence=(last['command']+'; '+str(line)) if last else 'Opt-in clean baseline, P09 and M12 attempts timed out at 90 seconds. No completed functional failure observed.')
 if t.endswith('LongMeasurements'):r['mutants_in_matrix']=1;r['matrix_rows']=[t];r['over_budget']=True;r['reason']='610 serial measurements cannot complete within 90 seconds; no completed mutation result or three-run median.'
 if verdict=='subsumed':r['subsumption_mutants']=len(kills)
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2))
summary={'origin_commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'nproc':5,'setup_seconds':0,'npm_seconds':json.loads((p/'initial.json').read_text())[0]['wall_seconds'],'row_test_seconds':seconds,'subsumer_seconds':famsec,'runtime_flags':'runtime native.Flags: release O2; sanitized O1 with address/undefined sanitizers','matrix_wall_seconds':sum(r['wall'] for r in matrix),'family_wall_seconds':sum(r['wall'] for r in family),'validation_wall_seconds':sum(r['wall'] for r in json.loads((p/'validation.json').read_text())),'timing_wall_seconds':sum(r['wall'] for r in json.loads((p/'timing.json').read_text())),'long_attempt_wall_seconds':sum(r['wall'] for r in json.loads((p/'long.json').read_text())) if (p/'long.json').exists() else None,'coverage_go_functions':[line for line in (p/'coverage-functions.txt').read_text().splitlines() if not line.endswith('0.0%') and not line.startswith('total:')]}
(p/'summary.json').write_text(json.dumps(summary,indent=2));table=[]
for m in plan['mutants']:
 r=next((r for r in matrix if r['id']==m['id']),None);table.append('| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+m['old'].replace('|','\\|')+'` -> `'+m['new'].replace('|','\\|')+'` | '+(', '.join(r['failures']+([fam] if r.get('family_failed') else [])) if r else 'not run')+' |')
(p/'mutants.md').write_text('| ID | Origin location | Change | Failed rows |\n|---|---|---|---|\n'+'\n'.join(table)+'\n')
