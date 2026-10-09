from pathlib import Path
import json,re,statistics,csv,subprocess
p=Path('review/test-audit/stage1-cohere-estree-deep');package='stage1/cohere/estree'
raw=['TestDeepGrammar','TestGeneratedAgreement','TestOriginalLibraries','TestDecoratedExports','TestDecoratedExportsPlantedDisagreement','TestDecoratedExportMutant','TestDecoratedExportMutantPlantedSurvivor','TestDecoratedExportLibraries','TestRecoveredExpressions','TestRecoveredExpressionMutant','TestUnattachedDecorator','TestUnattachedDecoratorControl']
prod=['TestDeepGrammar','TestGeneratedAgreement','TestDecoratedExports','TestRecoveredExpressions','TestUnattachedDecorator']; witnesses=['TestDecoratedExportsPlantedDisagreement','TestDecoratedExportMutant','TestDecoratedExportMutantPlantedSurvivor','TestRecoveredExpressionMutant']
family='TestOriginalLibraries family';fnames={'TestDeepGrammar':'deep_test.go','TestGeneratedAgreement':'estree_test.go','TestOriginalLibraries':'estree_test.go','TestDecoratedExports':'exports_test.go','TestDecoratedExportsPlantedDisagreement':'exports_test.go','TestDecoratedExportMutant':'exports_test.go','TestDecoratedExportMutantPlantedSurvivor':'exports_test.go','TestDecoratedExportLibraries':'exports_test.go','TestRecoveredExpressions':'expressions_test.go','TestRecoveredExpressionMutant':'expressions_test.go','TestUnattachedDecorator':'expressions_test.go','TestUnattachedDecoratorControl':'expressions_test.go'}
def events(f):
 out=[]
 for line in f.read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
med={};samples={}
for row in raw+['TestOriginalLibrariesFamily']:
 vals=[]
 for f in sorted(p.glob('timing-'+row+'-*.log')):
  for e in events(f):
   m=re.search(r'^ok\s+\S+\s+([\d.]+)s',e.get('Output',''))
   if m:vals.append(float(m[1]))
 assert len(vals)==3,(row,vals)
 key=family if row=='TestOriginalLibrariesFamily' else row;samples[key]=vals;med[key]=statistics.median(vals)
(p/'medians.json').write_text(json.dumps({'medians':med,'three_samples':samples},indent=2)+'\n')
runs=json.loads((p/'matrix-runs.json').read_text());builds=json.loads((p/'native-builds.json').read_text());valid={b['id'] for b in builds if b['exit']==0};matrix=[]
def status(path,row):
 es=events(path)
 if any('test timed out' in e.get('Output','') for e in es):return 'over-budget'
 done=[e['Action'] for e in es if e.get('Test')==row and e['Action'] in ['pass','fail','skip']]
 return done[-1] if done else 'unknown'
def line(path,row):
 candidates=[e.get('Output','').strip().splitlines()[0] for e in events(path) if (e.get('Test')==row or e.get('Test','').startswith(row+'/')) and re.search(r'\w+_test.go:\d+:',e.get('Output',''))]
 failures=[s for s in candidates if any(x in s for x in ['line ', 'mutant survived','planted failure','timeout=','expected','acceptance','did not','failed']) and not s.startswith('BUILD')]
 return failures[0] if failures else candidates[-1] if candidates else 'No failing line observed'
for mid in ['M1','M2','M3','M4']:
 results={row:status(p/(mid+'-'+row+'.log'),row) if (p/(mid+'-'+row+'.log')).exists() else 'not-reached' for row in prod}
 kills=[row for row,s in results.items() if s=='fail']
 matrix.append(dict(id=mid,valid_standalone_native_build=mid in valid,results=results,kills=kills,unique=kills if len(kills)==1 else []))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
with (p/'matrix.csv').open('w') as out:
 w=csv.writer(out);w.writerow(['mutant','native-build-valid']+prod)
 for m in matrix:w.writerow([m['id'],m['valid_standalone_native_build']]+[m['results'][r] for r in prod])
rows=[]
for name in raw:
 if name=='TestDecoratedExportLibraries':continue
 isfam=name=='TestOriginalLibraries';row=family if isfam else name
 obj=dict(test=row,package=package,file=[package+'/estree_test.go',package+'/exports_test.go'] if isfam else package+'/'+fnames[name],seconds=med[row],kills=[],unique_kills=[],last_proven_fail=None,subsumed_by=[],mutants_in_matrix=0,probe_kills=[],subsumer_seconds=None,vacuous=None,bounded=True,matrix_rows=prod if name in prod else [])
 if name in prod:
  obj['oracle_kind']='self' if name=='TestUnattachedDecorator' else 'external-run'
  obj['oracle']='Handwritten ESTree unattached decorator diagnostic substring, failed exit, empty stdout, and 2-second CPU deadline; no external authority checked.' if name=='TestUnattachedDecorator' else 'Unmodified Go cohere ESTree API through testdata/oracle.go. Exact canonical bytes are compared with the TypeScript port on source Node, sanitized native and emitted JavaScript.'
  k=[m['id'] for m in matrix if m['valid_standalone_native_build'] and name in m['kills']];unique=[m['id'] for m in matrix if m['valid_standalone_native_build'] and m['kills']==[name]]
  obj.update(kills=k,unique_kills=unique,mutants_in_matrix=4,mutants_run=[m['id'] for m in matrix if m['results'][name]!='not-reached'])
  if unique:obj['verdict']='slow-worthy' if med[name]>60 else 'sacred'
  elif k:
   candidates=[r for r in prod if r!=name and all(r in m['kills'] for m in matrix if m['id'] in k)]
   if candidates:
    chosen=min(candidates,key=lambda r:med[r]);obj.update(verdict='subsumed',subsumed_by=[chosen],subsumer_seconds=med[chosen],subsumption_kills=len(k))
   else:obj.update(verdict='overlapping',subsumed_by=sorted(set(r for m in matrix if m['id'] in k for r in m['kills'] if r!=name)),subsumption_kills=len(k))
  else:obj['verdict']='cannot-judge' if any(m['results'][name] in ['unknown','over-budget'] for m in matrix) else 'untrue'
  probe=p/('P1-'+name+'.log');ps=status(probe,name);obj['probe_status']=ps;obj['probe_kills']=['P1'] if ps=='fail' else [];obj['vacuous']=ps=='pass' if ps in ['pass','fail'] else None
  if k:
   mid=k[-1];r=next(x for x in runs if x['label']==mid+'-'+name);fail=line(p/(r['label']+'.log'),name);obj['last_proven_fail']=mid+': '+fail;obj['evidence']='selector file /tmp/u085/selector='+mid+'; '+r['command']+' > '+str(p/(r['label']+'.log'))+' 2>&1; '+fail
 elif name in witnesses:
  label='W1-planted' if 'Planted' in name else 'W1-'+name;path=p/(label+'.log');s=status(path,name);fail=line(path,name)
  obj.update(oracle_kind='self' if 'Planted' in name else ['external-run','self'] if name=='TestDecoratedExportMutant' else 'external-run',oracle='Planted agree/disagree answers and exact owning-shard failure are handwritten.' if 'Planted' in name else 'Go cohere bytes and expected disagreement from a built-in source mutant; production mutation failures do not establish the witness.',verdict='witness' if s=='fail' else 'untrue' if s=='pass' else 'cannot-judge',witness_checks=['W1'],evidence='timeout 120 go test -overlay=/tmp/u085/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^'+name+'$; '+fail)
  if 'Planted' in name:
   obj['evidence']='ADAMIC_ESTREE_LIBRARY=/tmp/u085/library timeout 120 go test -overlay=/tmp/u085/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^(TestDecoratedExportsPlantedDisagreement|TestDecoratedExportMutantPlantedSurvivor)$ > '+str(path)+' 2>&1; '+fail
  else:
   observed=next(x for x in runs if x['label']==label);obj['evidence']=observed['command']+' > '+str(path)+' 2>&1; '+fail
  if s=='fail':obj['last_proven_fail']='W1: '+fail
 elif name=='TestUnattachedDecoratorControl':
  path=p/'W2-TestUnattachedDecoratorControl.log';s=status(path,name)
  obj.update(oracle_kind=['external-run','self'],oracle='Go cohere audit must contain two errors; guard-disabled source Node and native must contain two Program records. Counts do not establish diagnostic identity or correct AST bytes. This proves the fixture precondition but does not execute the refusal comparison it is meant to support.',verdict='untrue' if s=='pass' else 'witness' if s=='fail' else 'cannot-judge',witness_checks=['W2'],evidence='timeout 120 go test -overlay=/tmp/u085/weak/W2.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestUnattachedDecoratorControl$; refusedBeforeDeadline disabled, --- PASS: TestUnattachedDecoratorControl')
 else:
  obj.update(members=['TestOriginalLibraries','TestDecoratedExportLibraries'],oracle_kind=['external-run','self'],oracle='Go cohere compared with pinned @typescript-eslint/typescript-estree 8.65.0 and Prettier 3.9.6 using TypeScript 6.0.3; exact handwritten deltas permit three documented whitespace gaps in the generated corpus.',verdict='cannot-judge',reason='No Adamic port code executes. A meaningful mutation would change one of the external oracles, which the brief prohibits.',evidence='ADAMIC_ESTREE_LIBRARY=/tmp/u085/library go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^(TestOriginalLibraries|TestDecoratedExportLibraries)$; three family runs pass, no failing line produced')
 rows.append(obj)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# Retain observed cold build messages separately from warm test medians.
cold=[]
for f in p.glob('*.log'):
 for e in events(f):
  text=e.get('Output','')
  if 'BUILD ' in text or 'build lowered/sanitized/emitted products:' in text:cold.append(dict(log=f.name,test=e.get('Test'),output=text.strip()))
(p/'cold-build-observations.json').write_text(json.dumps(cold,indent=2)+'\n')
base=json.loads((p/'baseline-runs.json').read_text());phases=dict(nproc=5,setup_seconds=0,library_install_seconds=3.989,whole_package_binary_seconds=90.022,scoped_package_binary_seconds=90.023,isolated_baseline_command_wall_seconds=sum(r['wall'] for r in base),isolated_baseline_binary_seconds=sum(sum(samples[r['row']]) for r in []),matrix_probe_witness_command_wall_seconds=sum(r['wall'] for r in runs),standalone_native_builds=builds,library_family_binary_seconds=sum(samples[family]))
phases['isolated_baseline_binary_seconds']=sum(sum(samples[r]) for r in raw)
(p/'phases.json').write_text(json.dumps(phases,indent=2)+'\n')
summary=['u085 starts at origin/main 12e77e8972a2e606cab6db05d84f428246a85339; all 12 names remain in the named files.','All 36 isolated clean runs pass with both original-library rows enabled; nproc 5.','The two original-library wrappers form one family, giving 11 judged rows.','Four fixed production mutants and one empty-entry probe; matrix bounded to reached port rows.','Evidence: test-audit/stage1-cohere-estree-deep, review/test-audit/stage1-cohere-estree-deep/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n| id | origin/main file:line | change | failed bounded rows |\n|---|---|---|---|\n'
menu=json.loads((p/'menu.json').read_text())
for m in menu:
 x=next(x for x in matrix if x['id']==m['id']);report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['kind']+' | '+', '.join(x['kills'])+' |\n'
survivors=[m['id'] for m in matrix if not m['kills']];report+='\nSurvivors: '+(', '.join(survivors) if survivors else 'none in the observed bounded matrix')+'. Kills outside the reached rows are unknown. No repo-wide replay was run.\n\nCode and oracle were named before production mutation in code-and-oracle.md. The named-function index and full reached ESTree function inventory, including anonymous callbacks, per-row reach sets and compressed V8 evidence are saved. The menu was finalized before any production mutant outcome. M2 was changed from child-index substitution to dropping the operator statement before execution, to avoid quadratic recursive conversion. The menu is four mutants because this is a native port; no supplemental production mutant was used.\n\nP1 returns an empty string from pipeline.answer. Only production rows calling that entry are probed. Witnesses and the library family have vacuous=null, because port probes do not judge their jobs. Production probes are separate from kills.\n\nWitness W1 makes firstDifference always return an empty difference. The two planted proofs and both actual wrong-output witnesses must fail under this weakened comparison. W2 disables refusedBeforeDeadline entirely. TestUnattachedDecoratorControl still passes; it checks that a built-in guard mutant accepts two inputs, but does not apply the acceptance/refusal comparison to prove it can catch those acceptances. Its untrue verdict is specifically the brief\'s witness-strength result, not a claim that its fixture precondition is useless. Neither experiment changes Go cohere or the pinned libraries.\n\nBrief ambiguities, costs and constraints:\n\n- The historical commit 8de93800f4 differs from fetched origin/main 12e77e8972. All requested names still exist; no name moved or vanished. Every mutant/probe line and standalone diff uses the starting commit.\n- The whole package cooks at 90 seconds after TestAcceptanceGrammar passes in 77.64 seconds. The requested combined slice also cooks at 90 seconds after deep grammar passes in 66.44 seconds. These are accumulated-budget timeouts, not red assertion baselines.\n- Deep grammar and generated agreement each take more than 60 seconds alone. Combining reached rows in one 90-second binary would hide later rows. Each bounded matrix cell therefore runs its row alone under the same timeout. All clean rows were observed three times before production mutation.\n- The older build(t) helper does not cache native products. A file-driven selector cannot prevent it rebuilding on every deep/generated invocation without a harness change. The cache-aware misc/recovery helpers share the stable switched source; no build harness was changed. The shared selector cache is correct for port-source mutants whose choice is read at runtime. Standalone validation gives each fixed source its own build cache.\n- Restored-source witness products were seeded into the selector cache by copying four exact content-addressed products built in this session; keys and paths are in witness-cache-seed.json. These do not match switched-source keys and cannot supply a production mutant answer.\n- The 30-minute port target competes with 36 required isolated runs, native matrix runs and standalone build checks. Native rebuilds account for most of the elapsed time; no individual observation was allowed beyond its 90-second budget.\n- Two library wrappers have the same checker and differ only by input recipe and expected gap count, so they are one family. Their native-port probes are inapplicable because they execute no Adamic source. Mutating Go cohere to force them to fail would mutate the oracle prohibited by the brief.\n- A built-in mutant witness is not a production coverage row. Weakening its comparison gives its verdict; other production failures or successful fixture preconditions do not. The unattached control demonstrates the distinction.\n- Go cohere decides exact agreement, while the orphan-refusal diagnostic is a handwritten port string. The library family also includes handwritten whitespace exceptions; it therefore has a mixed external-run/self oracle. No outside spec diagnostic was independently checked.\n- Pinned library dependencies were absent despite the warm toolchain. Installing them took 3.989 seconds. npm ci also ran in stage3/api even though these tests use plain Node without those modules.\n- Fixture extraction initially treated go run\'s -- separator as a source filename. It was corrected before coverage or production mutation; only the successful four corpus coverage files support reach claims.\n- Dropping M2\'s whole operator-field statement leaves operator used for choosing BinaryExpression versus LogicalExpression. Dropping M4\'s whole guard loop leaves no unused loop local. Standalone probe P1 replaces the complete answer body.\n- Repository log ignore rules require explicit force-adding the recorded logs; replay files are stored under review and scratch Go drivers use .go.txt so normal package discovery does not compile them.\n\nMeasured phases:\n\n```json\n'+json.dumps(phases,indent=2)+'\n```\n\nCold lowering and native build observations are in cold-build-observations.json; matrix-runs.json distinguishes command wall time from test-binary timing. No setup script ran because env.sh worked. Whole/scoped baselines were narrowed after 90 seconds, not allowed to continue.\n\nLimits: bounded uniqueness only, no other package run, no repo-wide uniqueness, no external-oracle mutation, no strengthening or rewriting of production tests, no PR or main push. The original-library family cannot receive a meaningful Adamic production-mutant verdict. Production source restoration and native-diff check results are part of the evidence.\n'
(p/'report.md').write_text(report)
print('Verdicts',[(r['test'],r['verdict'],r['seconds']) for r in rows])
print('Native builds',builds)
