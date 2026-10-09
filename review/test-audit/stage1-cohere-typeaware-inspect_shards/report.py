import pathlib,json,re,statistics,shlex,subprocess
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-typeaware-inspect_shards')
def events(name):
 f=p/name
 if not f.exists():return []
 out=[]
 for line in f.read_text().splitlines():
  try:out.append(json.loads(line))
  except ValueError:pass
 return out
def seconds(name):
 for e in reversed(events(name)):
  m=re.search(r'\tok?\s*',e.get('Output',''))
  if e.get('Action')=='pass' and not e.get('Test'):return e.get('Elapsed')
 return None
def fails(name):return [e['Test'] for e in events(name) if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']]
def fail_line(name,test):
 es=events(name)
 for e in es:
  if e.get('Test','').split('/')[0]==test and e.get('OutputType')=='error':return e['Output'].strip()
 for e in es:
  if e.get('Test','').split('/')[0]==test and '--- FAIL:' in e.get('Output',''):return e['Output'].strip()
 return None
scope=json.loads((p/'scope.json').read_text());family='TestSixRuleAgreementAndMutants family';names=['TestInspectRequestRefusals','TestShadowIndexMissingBinding_000','TestSharedProductPublication','TestSixBuildCallbackUsesProductDirectory','TestSixShardUnionRejectsLossAndDuplication','TestSixShardPlantedDisagreement',family,'TestSixRuleAgreementAndMutants_Setup','TestSixPinnedFlags','TestFactsDecoderGuards','TestTypeAwareUnitDeadline','TestTypeAwareNativeBuildChild','TestTypeAwareContextKillsCompilerGroup','TestPinnedTypeFlags']
commands={}
for filename in ['control-runs.json','extra-control-runs.json','production-runs.json','entry-probe-runs.json']:
 if (p/filename).exists():
  for x in json.loads((p/filename).read_text()):commands[x['id']]=' '.join(shlex.quote(s) for s in x['command'])+' > '+x['id']+'.log 2>&1'
commands['W5']='ADAMIC_BUILD_CACHE_DIR=/tmp/u145/cache/W5 ADAMIC_TYPESCRIPT_SOURCE=/tmp/u145/typescript ADAMIC_TYPEAWARE_BENCH=1 timeout 120 go test -overlay /tmp/u145/W5.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run ^TestShadowIndexMissingBinding_000$ > W5-fresh.log 2>&1'
controlled={'TestSharedProductPublication':('S1','setup-check'),'TestSixBuildCallbackUsesProductDirectory':('S2','setup-check'),'TestSixShardUnionRejectsLossAndDuplication':('W1','setup-check'),'TestSixShardPlantedDisagreement':('W2','witness'),'TestTypeAwareUnitDeadline':('W3','witness'),'TestTypeAwareContextKillsCompilerGroup':('W4','witness'),'TestSixRuleAgreementAndMutants_Setup':('S3','setup-check'),'TestShadowIndexMissingBinding_000':('W5','witness')}
rows=[]
for name in names:
 members=[f'TestSixRuleAgreementAndMutants_{i:03}' for i in range(95)]+['TestSixRuleAgreementAndMutantsUnion'] if name==family else [name]
 file=[scope['locations'][m] for m in members] if name==family else scope['locations'][name]
 obs=[seconds(f'timing-{name}-{i}.log') for i in range(3)]
 if name in ['TestInspectRequestRefusals','TestShadowIndexMissingBinding_000']:obs=[seconds(f'baseline-{name}.log')]+[seconds(f'heavy-timing-{name}-{i}.log') for i in range(2)]
 if name=='TestSixRuleAgreementAndMutants_Setup':obs=[seconds(f'heavy-timing-{name}-{i}.log') for i in range(2)]+[seconds('heavy-timing-setup-third.log')]
 sec=statistics.median(obs) if all(x is not None for x in obs) else None
 kills=[]
 for id in ['M1','M2','M3','M4']:
  actual='TestSixRuleAgreementAndMutants_000' if name==family else name
  if actual in fails(id+'.log'):kills.append(id)
 oracle='Hand-written expected construction results, errors or process markers'
 kind='self'
 if name==family:oracle='Exact stdout bytes from Go cohere; sanitizer stderr and exits; hand-written reporting controls. Only reached leaf 000 was replayed under production mutants.';kind=['external-run','self']
 if name=='TestInspectRequestRefusals':oracle='Self-written exit 0 validity control and exit 70 plus named refusal substrings. Validity does not check returned facts or their count.'
 if name=='TestShadowIndexMissingBinding_000':oracle='Self-written exit 70 and missing binding index substring for a native built-in mutant; fresh guard weakening proves the witness.'
 if name=='TestFactsDecoderGuards':oracle='Self-written valid stdout 64 plus exit 70 and named malformed-frame errors. Strict/present boolean inversion M2 passes this row.'
 if name in ['TestSixPinnedFlags','TestPinnedTypeFlags']:oracle='Pinned Go cohere TypeScript checker TypeFlags constants. Checked union=1<<27=134217728 and mask sum=334017 against source and successful tests. These rows do not execute port constants.';kind='external-authority'
 verdict='cannot-judge';control=[];last=None;evidence=None
 if name in controlled:
  id,verdict=controlled[name];log='W5-fresh.log' if id=='W5' else id+'.log';line=fail_line(log,name)
  if name not in fails(log):verdict='cannot-judge'
  else:control=[id];last=id+': '+str(line);evidence=commands.get(id,id)+'; '+str(line)
 if name==family:verdict='sacred' if 'M2' in kills else 'cannot-judge'
 if name=='TestInspectRequestRefusals':verdict='sacred' if 'M4' in kills else 'cannot-judge'
 if name=='TestFactsDecoderGuards':verdict='subsumed' if kills and all(k in ['M1','M3'] for k in kills) else 'cannot-judge'
 if name=='TestTypeAwareNativeBuildChild':verdict='helper';oracle='Subprocess native builder, no independent oracle'
 if kills:
  id=kills[-1];actual='TestSixRuleAgreementAndMutants_000' if name==family else name;line=fail_line(id+'.log',actual);last=id+': '+str(line);evidence=commands[id]+'; '+str(line)
 probes=[]
 for id in ['P1','P2','P3']:
  targets=['TestSixRuleAgreementAndMutants_000','TestSixRuleAgreementAndMutants_033'] if name==family else [name]
  if any(t in fails(id+'.log') for t in targets) and not (name==family and id=='P1'):probes.append(id)
  if name==family and id=='P2' and 'TestSixRuleAgreementAndMutants_033' in fails('P2-cost-alone.log') and id not in probes: probes.append(id)
 own_probed=name in ['TestFactsDecoderGuards','TestInspectRequestRefusals',family]
 r={'test':name,'package':'stage1/cohere/typeaware','file':file,'seconds':sec,'oracle':oracle,'oracle_kind':kind,'kills':kills,'unique_kills':['M2'] if name==family and 'M2' in kills else ['M4'] if name=='TestInspectRequestRefusals' and 'M4' in kills else [],'last_proven_fail':last,'verdict':verdict,'subsumed_by':[family] if verdict=='subsumed' else [],'mutants_in_matrix':4 if name==family else 3 if name=='TestFactsDecoderGuards' else 1 if name=='TestInspectRequestRefusals' else 0,'probe_kills':probes,'subsumer_seconds':None,'vacuous':False if own_probed and probes else None,'bounded':True,'matrix_rows':['TestFactsDecoderGuards','TestInspectRequestRefusals','TestSixRuleAgreementAndMutants_000','TestSixRuleAgreementAndMutants_033'] if name in ['TestFactsDecoderGuards','TestInspectRequestRefusals',family] else [name],'evidence':evidence,'control_kills':control,'timing_observations':obs}
 if name==family:r['members']=members;r['timing_status']='Complete family timed out at 90 s; three successful complete-family observations unavailable.'
 if verdict=='subsumed':r['subsumption_basis_mutants']=len(kills);r['subsumer_timing_status']='Complete family exceeded 90 s; no valid median.'
 if name in ['TestSixPinnedFlags','TestPinnedTypeFlags']:r['reason']='A mutation of the Go constants would change the authority; no permitted port mutation can reach these assertions.'
 if verdict=='helper':r['parents']=['TestInspectRequestRefusals','TestShadowIndexMissingBinding_000',family,'TestSixRuleAgreementAndMutants_Setup'];r['seconds']=None
 if name=='TestInspectRequestRefusals':r['vacuous_subcases']=[e['Test'] for e in events('P2.log') if e.get('Action')=='pass' and e.get('Test','').startswith(name+'/') and not e['Test'].endswith('/shard-001')]
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
matrix=[]
for id in ['M1','M2','M3','M4','P1','P2','P3','S1','S2','S3','W1','W2','W3','W4','W5']:
 log='W5-fresh.log' if id=='W5' else id+'.log';es=events(log);es = es + (events('P2-cost-alone.log') if id=='P2' else [])
 statuses={e['Test']:e['Action'] for e in es if e.get('Test') and e.get('Action') in ['pass','fail','skip']};matrix.append({'id':id,'kind':'production' if id.startswith('M') else 'probe' if id.startswith('P') else 'control','log':log,'observed':statuses,'not_observed':'unknown, not passing'})
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
skips={f.name:[e['Test'] for e in events(f.name) if e.get('Action')=='skip' and e.get('Test')] for f in p.glob('*.log')};(p/'skips.json').write_text(json.dumps({k:v for k,v in skips.items() if v},indent=2)+'\n')
print([(r['test'],r['verdict'],r['seconds'],r['kills'],r['probe_kills']) for r in rows])
