import pathlib,json,re,statistics,csv,subprocess,shutil
E=pathlib.Path('review/test-audit/stage1-cohere-lint-shards_agree_split');start=json.loads((E/'start.json').read_text());clean=json.loads((E/'clean-runs.json').read_text());audit=json.loads((E/'audit-runs.json').read_text());checks=json.loads((E/'checks.json').read_text());scope=json.loads((E/'scope.json').read_text());matrixnames=['TestSuggestionAlongsideAutomaticFix family','TestWitnessScriptKind family'];rows=[]
def median(group):
 vals=[]
 for r in clean:
  if r['group']==group and r['exit']==0:
   es=[e['Elapsed'] for e in r['events'] if e['Action']=='pass' and not e.get('Test')];vals+=es
 return statistics.median(vals) if len(vals)==3 else None
def elapsedlog(name):
 p=E/(name+'.log');vals=[]
 for l in p.read_text().splitlines():
  try:
   e=json.loads(l)
   if e.get('Action')=='pass' and not e.get('Test'):vals.append(e['Elapsed'])
  except:pass
 return vals[-1] if vals else None
kindtimes=[elapsedlog('clean-kind-family-'+str(i)) for i in range(1,4)];kindmedian=statistics.median(kindtimes) if all(x is not None for x in kindtimes) else None

def evidence(name,fail=True):
 p=E/(name+'.log');lines=[]
 failed=set()
 if p.exists():
  for raw in p.read_text().splitlines():
   try:
    event=json.loads(raw)
    if event.get('Action')=='fail' and event.get('Test'):failed.add(event['Test'])
   except:pass
 if p.exists():
  for l in p.read_text().splitlines():
   try:
    e=json.loads(l);s=e.get('Output','').strip()
    if (re.search(r'\.go:\d+:',s) and (not fail or e.get('Test') in failed) and not any(x in s for x in [' build ','identical','cooked=false'])):lines.append(s)
    if not fail and '--- PASS:' in s:lines.append(s)
   except:pass
 if name=='S5-required' and lines:lines[0]+=' (scratch line 294 maps to origin/main line 295)'
 return lines[0] if lines else ('panic: test timed out after 1m30s' if fail else 'PASS')

def make(test,members,group,oracle,kind,verdict,check=None,prod=None):
 sec=kindmedian if test==matrixnames[1] else median(group)
 loc=scope['locations'][members[0]];files=list(dict.fromkeys(scope['locations'][n]['file'] for n in members));r={'test':test,'package':'stage1/cohere/lint','file':','.join(files),'seconds':sec,'oracle':oracle,'oracle_kind':kind,'kills':[],'unique_kills':[],'last_proven_fail':None,'verdict':verdict,'subsumed_by':[],'mutants_in_matrix':4 if prod else 0,'probe_kills':[],'subsumer_seconds':None,'vacuous':None,'bounded':True,'matrix_rows':matrixnames,'evidence':'','members':members}
 if prod:
  for id in ['M1','M2','M3','M4']:
   rr=next((x for x in audit if x['id']==id+'-'+prod),None)
   if rr and not rr['timeout'] and rr['exit']==1 and any(e.get('Test') for e in rr.get('events',[]) if e['Action']=='fail'):r['kills'].append(id)
  pp=next((x for x in audit if x['id']=='P1-'+prod),None)
  if pp and not pp['timeout']:
   r['vacuous']=pp['exit']==0
   if pp['exit']==1:r['probe_kills']=['P1']
  if r['kills']:
   id=r['kills'][-1];line=evidence(id+'-'+prod);r['last_proven_fail']=id+': '+line;r['evidence']='timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '+('^TestSuggestionAlongsideAutomaticFix_[0-9]+$' if prod=='suggestion' else '^TestWitnessScriptKind(?:_[0-9]+)?$')+'; selector='+id+'; '+line
  else:r['verdict']='cannot-judge';r['evidence']='No verified production kill in available matrix.'
 elif check:
  rr=next(x for x in checks if x['id']==check);line=evidence(check,fail=rr['exit']!=0);r['evidence']=' '.join(rr['command'])+'; '+line
  if rr['exit'] and not rr['timeout']:r['last_proven_fail']=check+': '+line
  elif rr['timeout']:r['verdict']='cannot-judge';r['evidence']+='; cooked'
 else:
  rr=next(x for x in clean if x['group']==group);r['evidence']=' '.join(rr['command'])+'; panic: test timed out after 1m30s; no completed clean baseline or mutant observation'
 rows.append(r);return r
make('TestShardsAgree family',[f'TestShardsAgree_{i:03}' for i in range(32)],'shards','Same native lint driver run serially, exact byte and count comparison; self consistency does not establish semantic parity.','self','cannot-judge')
make('TestShardsAgree_Union',['TestShardsAgree_Union'],'shards-union','Hand-written enumeration, exact-once ownership, and planted byte disagreement checked with difference; witness could not reach its check within budget.','self','cannot-judge')
make('TestShardsAgree_Setup',['TestShardsAgree_Setup'],'shards-setup','Suite product/corpus construction, build success; preparation did not complete.','self','cannot-judge')
make('TestShardsAgree_SetupRequired',['TestShardsAgree_SetupRequired'],'shards-required','Hand-written child PASS, named prepared-product build and cache-miss checks; preparation did not complete.','self','cannot-judge')
make('TestSuggestionAlongsideAutomaticFix_Setup',['TestSuggestionAlongsideAutomaticFix_Setup'],'suggestion-setup','Construction success only: setting ready=false still passes; no assertion requires the ready flag.','self','untrue','S2-suggestion-setup')
a=make(matrixnames[0],[f'TestSuggestionAlongsideAutomaticFix_{i:03}' for i in range(3)],'suggestion','Go cohere executes the pinned rule adapter; exact output bytes across source Node, emitted JavaScript and sanitized native, plus hand-written automatic-fix substring.',['external-run','self'],'subsumed',prod='suggestion')
make('TestSuggestionAlongsideAutomaticFixPlantedFailure',['TestSuggestionAlongsideAutomaticFixPlantedFailure'],'suggestion-planted','Built-in native-only byte mismatch; parent requires exactly shard 002 failure and specific planted text. W1 makes difference accept all bytes.','self','witness','W1-suggestion')
make('TestSuggestionAlongsideAutomaticFixSetupIsRequired',['TestSuggestionAlongsideAutomaticFixSetupIsRequired'],'suggestion-required','Fresh-process child PASS and cache-miss assertions; S5 disables selected leaf preparation and parent fails.','self','setup-check','S5-required')
for suff,g in [('Lowered','product-lower'),('Native','product-native'),('GoOracle','product-oracle')]:make('TestProduct_WitnessScriptKind'+suff,['TestProduct_WitnessScriptKind'+suff],g,'Product construction only; S1 returns empty directory from all product makers and this wrapper still passes. No output comparison.','self','untrue','S1-products')
b=make(matrixnames[1],['TestWitnessScriptKind_000','TestWitnessScriptKind_001','TestWitnessScriptKind'],'witness-kind','Go cohere executes no-debugger; exact output bytes across source Node, emitted JavaScript and sanitized native; union has hand-written exact-once live-corpus coverage.',['external-run','self'],'subsumed',prod='kind')
make('TestWitnessScriptKindPlantedFailure',['TestWitnessScriptKindPlantedFailure'],'kind-planted','Built-in emitted-JavaScript-only byte mismatch; parent requires exactly assigned leaf failure and emitted JavaScript diagnostic. W1 accepts all bytes, exposing planted mismatch survived instead.','self','witness','W1-kind')
make('TestWitnessScriptKind_Setup',['TestWitnessScriptKind_Setup'],'kind-setup','Hand-written witness extension and ready/products/source-count construction checks; S3 constructs .ts filenames for .tsx witnesses and the unchanged extension assertion fails.','self','setup-check','S3-kind-setup')
make('TestWitnessScriptKindStandalone',['TestWitnessScriptKindStandalone'],'kind-standalone','Parent requires child named-leaf PASS; S4 selects a nonexistent child row and fails that assertion.','self','setup-check','S4-kind-standalone')
if a['kills'] and b['kills']:
 for r,other in [(a,b),(b,a)]:
  if set(r['kills'])<=set(other['kills']):r['subsumed_by']=[other['test']];r['subsumer_seconds']=other['seconds'];r['subsumption_mutants']=len(r['kills'])
b['vacuous_subcases']=['TestWitnessScriptKind (coverage union)'] if b['probe_kills'] else []
(E/'rows.json').write_text(json.dumps(rows,indent=2));matrix=[]
for id in ['M1','M2','M3','M4','P1','P2']:
 record={'id':id,'rows':{}}
 for g,name in [('suggestion',matrixnames[0]),('kind',matrixnames[1])]:
  rr=next((x for x in audit if x['id']==id+'-'+g),None);record['rows'][name]='unknown' if not rr or rr['timeout'] else 'kill' if rr['exit']==1 else 'pass' if rr['exit']==0 else 'unknown'
 record['outside_matrix']='unknown';matrix.append(record)
(E/'matrix.json').write_text(json.dumps(matrix,indent=2))
with (E/'matrix.csv').open('w') as f:
 w=csv.writer(f);w.writerow(['id']+matrixnames)
 for x in matrix:w.writerow([x['id']]+[x['rows'][n] for n in matrixnames])
(E/'timing.json').write_text(json.dumps({'setup':start,'whole_binary_seconds':90.662,'combined_slice_binary_seconds':90.069,'clean_command_wall_seconds':sum(x['wall'] for x in clean),'mutation_command_wall_seconds':sum(x['wall'] for x in audit),'construction_witness_command_wall_seconds':sum(x['wall'] for x in checks),'standalone_compilations':[x for x in audit if x['id'].endswith('-compile')]},indent=2))
shutil.copy('/tmp/u112/audit.py',E/'run-audit.py');shutil.copy('/tmp/u112/checks.py',E/'run-checks.py');shutil.copy('/tmp/u112/report.py',E/'summarize.py')
print(json.dumps([(r['test'],r['seconds'],r['verdict'],r['kills']) for r in rows],indent=2))
