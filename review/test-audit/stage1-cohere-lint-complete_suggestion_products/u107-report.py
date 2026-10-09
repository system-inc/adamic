import pathlib,json,re,statistics,subprocess,gzip,shutil
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-lint-complete_suggestion_products'
def records(label):
 f=p/(label+'.log')
 if not f.exists():return []
 a=[]
 for l in f.read_text(errors='replace').splitlines():
  try:a.append(json.loads(l))
  except:pass
 return a
def failures(label):
 a=[r['Test'] for r in records(label) if r.get('Action')=='fail' and r.get('Test') and '/' not in r['Test']]
 if label=='P2' and not a and any('panic:' in r.get('Output','') for r in records(label)):a=['TestChildCPUWaitGuard']
 return a
def elapsed(label):
 for r in records(label):
  m=re.search(r'^ok\s+\S+\s+(\d+(?:\.\d+)?)s',r.get('Output',''))
  if m:return float(m[1])
 return None
def failureline(label,members):
 needles={'S0':'missing field','S1':'shared preparation failed','S2':'build complete-suggestion-native:','S3':'exclusion exceeded','W1':'wrong planted failure','W2':'second suggestion edit mutant survived','W3':'loop guard:','W4':'backstop:','M4':': port ','M5':'wall backstop exceeded'}
 for r in records(label):
  if r.get('Test') in members and needles.get(label,'__NO_MATCH__') in r.get('Output',''):
   return r['Output'].strip().splitlines()[0]
 for r in records(label):
  if r.get('Test') in members and r.get('Action')=='output':
   s=r.get('Output','').strip()
   if re.search(r'\.go:\d+:',s) and not re.search(r'build complete-suggestion-\S+ \S+ (hit|miss) ',s) and not any(x in s for x in ['TestProduct_CompleteSuggestion:','shared preparation:','caught on','shard 00','compiler corpus stage1:']):return s.splitlines()[0]
 for r in records(label):
  if 'panic:' in r.get('Output',''):return r['Output'].strip().splitlines()[0]
 return 'No completed failing assertion observed'
rows=json.loads((p/'rows.json').read_text());aliases=['P','F','A','MW','S','PF','C','H','CH','WB','CW','E'];rowmap=dict(zip(aliases,[a[0] for a in rows]));portrows=['P','F','A','MW','S','C','H','CH','WB','CW']
meta=[('self','Build success and artifact paths only; no comparison of lint answers.','S2','setup-check'),('self','Hand-written oracle field substrings; checks captured Go fixture output, not port output.','S0','setup-check'),(['external-run','self'],'Go cohere complete serialized bytes versus source Node, emitted JavaScript and sanitized native; union has a synthetic ownership assertion.','M4','sacred'),('external-run','Go cohere bytes must differ from the built-in wrong second-edit endpoint. W2 forces that comparison to claim agreement.','W2','witness'),('self','Shared products must be prepared and readiness recorded.','S1','setup-check'),('self','Exactly shard 001 must fail and contain planted suggestion disagreement.','W1','witness'),('self','Exact hand-written retained paths and excluded folder names in a temporary git repository; git does not supply the expected answer.','S3','setup-check'),('self','Dormant subprocess adapter; returns immediately without its parent environment. No independent assertion.','', 'helper'),('self','Node infinite-loop child must return a named CPU-guard error. W3 suppresses the failure report, while preserving the CPU limit.','W3','witness'),('self','Sleeping Node child must return a named wall-backstop error. W4 suppresses the failure report, while preserving the kill.','W4','witness'),('external-run','Node CPU/wait/CPU child must exit successfully, wall time exceed 1 second and CPU remain below 1 second; no stdout comparison.','M5','sacred'),(['external-run','self'],'Go cohere bytes compared with source Node, emitted JavaScript and native; planted child must fail with the specific mismatch text.','W5','witness')]
testfiles=list((root/'stage1/cohere/lint').glob('*_test.go'));locations={}
for f in testfiles:
 s=subprocess.check_output(['git','show','16f436a1:'+str(f.relative_to(root))],text=True)
 for m in re.finditer(r'^func (Test\w+)\(',s,re.M):locations[m[1]]=str(f.relative_to(root))+':'+str(s[:m.start()].count('\n')+1)
result=[]
for i,((name,members),(kind,oracle,last,verdict)) in enumerate(zip(rows,meta)):
 times=[elapsed(f'timing-{i}-{j}') for j in range(3)]
 if i==11:times=[elapsed('emitted-individual-baseline'),elapsed('emitted-timing-0'),elapsed('emitted-timing-1')]
 seconds=statistics.median(times) if all(x is not None for x in times) else None
 validlast=bool(last and (p/(last+'-time.json')).exists() and any(x in failures(last) for x in members))
 if i==11 and not validlast:verdict='cannot-judge';last='';oracle+=' Preparation did not complete within the individual 90-second budget; witness weakening was not run.'
 kills=['M1','M2','M4'] if i==2 else ['M5'] if i==10 else []
 kills=[x for x in kills if any(y in failures(x) for y in members)]
 if i in [2,10] and not kills:verdict='cannot-judge'
 probes=['P1B'] if i==2 and any(x in failures('P1B') for x in members) else ['P2'] if i==10 and failures('P2') else []
 line=failureline(last,members) if validlast else None
 cmd=json.loads((p/(last+'-time.json')).read_text())['command'] if validlast else 'go test -list . ./stage1/cohere/lint/'
 if i==11 and not validlast:cmd='timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestEmittedJavaScriptMismatch_000$';line=failureline('emitted-individual-baseline',members)
 o=dict(id=aliases[i],test=name,members=members,package='stage1/cohere/lint',file=[locations[m] for m in members],seconds=seconds,timing_samples=times,oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=kills,last_proven_fail=(last+': '+line if validlast else None),verdict=verdict,subsumed_by=[],mutants_in_matrix=(4 if i in [0,1,2,4,6,7,8,9] else 2 if i==10 else 0),probe_kills=probes,subsumer_seconds=None,vacuous=(False if probes else None),bounded=True,matrix_rows=(['CW'] if i==10 else portrows if i not in [5,11] else []),evidence=cmd+'; '+(line or 'No independent failing assertion required for helper.'))
 if i==7:o['parent']='compilerShardLauncher adapter in lint_test.go:408; git grep found no caller at the starting commit'
 result.append(o)
(p/'report.json').write_text(json.dumps(result,indent=2,ensure_ascii=False))
(p/'matrix.json').write_text(json.dumps({'row_aliases':rowmap,'production':{m:{'failed_tests':failures(m),'failed_rows':[a for a,(n,ms) in zip(aliases,rows) if set(ms)&set(failures(m))]} for m in ['M1','M2','M3','M4','M5','M6']},'probes':{m:{'failed_tests':failures(m),'valid_compilation':any(r.get('Action')=='pass' and r.get('Test')=='TestProduct_CompleteSuggestionNative' for r in records(m)) if m=='P1B' else (p/'P2-vet.log').exists()} for m in ['P1B','P2']},'checks':{m:failures(m) for m in ['S0','S1','S2','S3','W1','W2','W3','W4','W5']},'unknown':'All other package rows and repo-wide kills. Witness production precondition failures never support verdicts.'},indent=2))
validation={}
for f in sorted(p.glob('*.diff')):
 check=subprocess.run(['git','apply','--cached','--check',str(f.relative_to(root))],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
 validation[f.name]={'applies_to_clean_starting_index':check.returncode==0,'output':check.stdout}
(p/'apply-checks.json').write_text(json.dumps(validation,indent=2));assert all(x['applies_to_clean_starting_index'] for x in validation.values())
# Save native build measurements exactly as logged, not inferred compile durations.
builds={}
for m in ['M1','M2','M3','M4','P1B']:
 a=[]
 for r in records(m):
  x=re.search(r'build (complete-suggestion-\S+) \S+ miss (\d+(?:\.\d+)?)',r.get('Output',''))
  if x:a.append({'product':x[1],'seconds':float(x[2])})
 builds[m]=a
command_wall=sum(json.loads(f.read_text()).get('seconds',0) for f in p.glob('*-time.json'))
(p/'timing-summary.json').write_text(json.dumps({'setup_seconds':0,'nproc':5,'npm_ci_seconds':0.768,'native_and_lowering_build_measurements':builds,'measured_go_test_command_wall_seconds':command_wall,'note':'Build product measurements are nested in command timings; do not sum them with command wall seconds. Go compilation/launch overhead is not separately isolated. Reading, source editing, vet and publication are outside the summed go-test timings.'},indent=2))
for source in ['/workspace/u107-run.py','/workspace/u107-extra.py','/workspace/u107-last.py','/workspace/u107-report.py']:shutil.copy(source,p/pathlib.Path(source).name)
print(json.dumps([(o['id'],o['seconds'],o['verdict'],o['last_proven_fail']) for o in result],ensure_ascii=False,indent=2))
