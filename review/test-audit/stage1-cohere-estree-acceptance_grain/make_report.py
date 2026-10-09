import json,statistics,shlex,collections
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-estree-acceptance_grain')
def records(name):
 result=[]
 for line in (p/name).read_text().splitlines():
  try:result.append(json.loads(line))
  except json.JSONDecodeError:pass
 return result
commands=records('commands.jsonl'); experiments=records('experiment-menu.jsonl'); timings=json.loads((p/'timings.json').read_text()); scope={r['test']:r for r in json.loads((p/'scope.json').read_text())}
cmd={r['id']:r for r in commands if 'id' in r}
results={r['id']:records(r['id']+'.log') for r in experiments}
def failed(id):return sorted({r['Test'] for r in results[id] if r.get('Action')=='fail' and r.get('Test') and '/' not in r['Test']})
def evidence(id,members):
 outputs=[r.get('Output','').strip() for r in results[id] if r.get('Test') in members]
 preferred=[x for x in outputs if x and not x.startswith(('===','---')) and ('OutputType'=='error' or any(s in x for s in ['survived','plant caught','must fail','No such','no such','line 1:','line 25:','timeout=false','build estree','build deep']))]
 # Prefer actual Go test error frames.
 errors=[r.get('Output','').strip() for r in results[id] if r.get('Test') in members and r.get('OutputType')=='error']
 if errors:line=errors[0]
 elif preferred:line=preferred[-1]
 else:line=next((x for x in outputs if x.startswith(('--- PASS','--- SKIP','--- FAIL'))),'no row output')
 command=shlex.join(cmd[id]['command']);return command+' > '+id+'.log 2>&1; '+line,line
production=['M1','M2','M3','M4'];matrix=['TestAcceptanceGrammar','TestAcceptanceDiagnostics']
names=['TestProduct_AcceptanceOracle','TestProduct_Acceptance family','TestAcceptanceMutants family','TestAcceptanceMutantsPlanted family','TestAcceptanceGrammar','TestAcceptanceDiagnostics','TestAcceptanceDiagnosticControl','TestRepositoryAgreement','TestCorpusNativeRefusals','TestDeadlineChild','TestDeepMutants_Setup','TestDeepMutants family']
oracles=[
 ('Successful Go-oracle product construction; self-written builder error checks, no artifact existence/content assertion.','self'),
 ('Successful lowered/native product construction; self-written builder error checks, returned paths are not inspected.','self'),
 ('Live Go cohere canonical bytes against built-in port mutants on source Node and sanitized native; self-written mutant-must-differ and shard-union checks.',['external-run','self']),
 ('Real child acceptance check using Go cohere; self-written expectation that only shard 000 rejects the planted surviving native result.',['external-run','self']),
 ('Live Go cohere canonical bytes compared exactly with source Node, sanitized native and emitted JavaScript.','external-run'),
 ('Live Go cohere rejects eight inputs; port check requires nonzero exit, empty stdout, no CPU timeout and ESTree parser substring. M4 preserves these while losing the specific diagnostic.','external-run'),
 ('Live Go cohere refusal plus self-written 0 Program substring for the validation-disabled port on source Node/native. The refusal checker it claims to witness is never invoked.',['external-run','self']),
 ('Frozen Go cohere canonical answers, not checked against Go this session because original snapshot/answers are absent.','external-authority'),
 ('Frozen Go cohere refusal statuses plus self-written panic/empty-stdout/deadline checks; specific refusal reasons are not compared. Frozen values not checked this session.',['external-authority','self']),
 ('Subprocess entry only; active when parent sets ADAMIC_ESTREE_DEADLINE_CHILD=1.','self'),
 ('Successful frozen-fixture construction; self-written builder error checks, no want-file validation.','self'),
 ('Live Go cohere bytes compared to three built-in port mutants on Node/native; same byte predicate also guards planted disagreement and shard ownership.',['external-run','self'])]
rows=[]
for i,t in enumerate(timings):
 members=t['members'];ks=[id for id in production if set(members)&set(failed(id))] if i in [4,5] else []
 unique=[id for id in ks if len(failed(id))==1]
 r={'test':names[i],'members':members,'package':'stage1/cohere/estree','file':scope[members[0]]['location'][0],'seconds':t['seconds'],'oracle':oracles[i][0],'oracle_kind':oracles[i][1],'kills':ks,'unique_kills':unique,'last_proven_fail':None,'verdict':'cannot-judge','subsumed_by':[],'mutants_in_matrix':4 if i in [4,5] else 0,'probe_kills':[],'subsumer_seconds':None,'vacuous':None,'bounded':True,'matrix_rows':matrix if i in [4,5] else members,'evidence':'','timing_runs':t['runs']}
 if i in [4,5]:
  r['verdict']='sacred' if unique else 'untrue'; r['vacuous']=not bool(set(members)&set(failed('P1')))
  r['probe_kills']=['P1'] if not r['vacuous'] else []
  id=ks[-1];r['evidence'],line=evidence(id,members);r['last_proven_fail']=id+': '+line
 elif i in [0,1,10]:
  id={0:'S4',1:'S5',10:'S6'}[i];r['verdict']='setup-check' if set(members)&set(failed(id)) else 'untrue'
  r['evidence'],line=evidence(id,members);r['last_proven_fail']=id+': '+line if r['verdict']=='setup-check' else None
  r['setup_kills']=[id] if r['last_proven_fail'] else [];r['construction_survivors']=[{0:'S1',1:'S2',10:'S3'}[i]]
 elif i in [2,3,11]:
  id='W2' if i==3 else 'W1';r['verdict']='witness' if set(members)&set(failed(id)) else 'untrue';r['evidence'],line=evidence(id,members)
  r['last_proven_fail']=id+': '+line if r['verdict']=='witness' else None;r['witness_kills']=[id] if r['last_proven_fail'] else []
  r['matrix_rows']=next(e['rows'] for e in experiments if e['id']==id)
 elif i==6:
  r['verdict']='witness' if failed('W3') else 'untrue';r['evidence'],line=evidence('W3',members);r['weakened_check_survivors']=['W3']
 elif i in [7,8]:
  r['evidence']='go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^'+members[0]+'$ > timing-'+str(i)+'-0.log 2>&1; '+next(x.get('Output','').strip() for x in records('timing-'+str(i)+'-0.log') if 'corpus' in x.get('Output','').lower() and ': ' in x.get('Output',''))
  r['reason']='Completed frozen corpus unavailable; all three runs skipped. Seconds measure only skipped test-binary invocation.';r['matrix_rows']=[]
 elif i==9:
  r['verdict']='helper';r['parents_in_unit']=['TestAcceptanceDiagnostics','TestCorpusNativeRefusals'];r['evidence']='go test -json -count=1 ./stage1/cohere/estree/ -run ^TestDeadlineChild$ > timing-9-0.log 2>&1; --- PASS: TestDeadlineChild (0.00s), dormant entry; production diagnostics invoke the active child.';r['matrix_rows']=[]
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
summary=['u084 audited origin/main '+(p/'base.txt').read_text().strip()+'.','All 20 names exist; family grouping yields 12 rows; four deep-mutant functions moved files.','Verdicts: '+', '.join(str(n)+' '+v for v,n in collections.Counter(r['verdict'] for r in rows).items())+'.','Four production mutants: three killed, one diagnostic-detail survivor; P1 killed both production rows.','Evidence: test-audit/stage1-cohere-estree-acceptance_grain, review/test-audit/stage1-cohere-estree-acceptance_grain/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'
report+='| ID | File:line at starting commit | Change | Failed rows |\n|---|---|---|---|\n'
for e in experiments:
 if e['kind']=='production':report+='| '+e['id']+' | '+e['file']+':'+str(e['line'])+' | `'+e['from']+'` -> `'+e['to']+'` | '+', '.join(failed(e['id']))+' |\n'
report+='\nM4 survivor: '+(p/'M4-witness-before.stderr').read_text().splitlines()[0]+' -> '+(p/'M4-witness-after.stderr').read_text().splitlines()[0]+'. Both direct Node runs have the same nonzero exit; both production rows pass. This leaves diagnostic detail unguarded in this bounded matrix.\n\n'
report+='P1 is an empty-answer probe, never a mutant kill. W1 disables firstDifference; W2 disables the mutant-survival comparison; W3 bypasses the refusal checker. S1/S3 omit artifact writes; S2 returns nonexistent product paths. S4/S5/S6 route artifact writes into missing subdirectories. Every experiment has a standalone diff and log.\n\n'
report+='| Harness/probe ID | Starting location | Observed failed functions |\n|---|---|---|\n'
for e in experiments:
 if e['kind']!='production':report+='| '+e['id']+' | '+e['file']+':'+str(e['line'])+' | '+(', '.join(failed(e['id'])) or 'none')+' |\n'
report+='\n'+(p/'issues.txt').read_text()+'\n\n'
report+='Warm setup skipped. '+(p/'toolchain.log').read_text().replace('\n','; ')+' npm ci reported 449 ms. Whole baseline: 90.025 s, timed out after grammar passed in 58.76 s; no preceding test failure. Bounded clean baseline: 52.116 s. All bounded runs use ADAMIC_NATIVE_SPLIT=1 and ADAMIC_NATIVE_JOBS=5.\n\n'
report+='Native rebuild-inclusive test times are upper bounds, not isolated clang measurements: '\
 +', '.join(id+' '+str(next(r['Elapsed'] for r in results[id] if r.get('Test')=='TestAcceptanceGrammar' and r['Action'] in ['pass','fail']))+' s' for id in production+['P1'])+'. Exact cold product build lines are in timings.json.\n\n'
report+='Recorded timing command wall total: '+str(round(sum(r['wall'] for r in commands if r['kind']=='timing'),3))+' s. Experiment command wall total: '+str(round(sum(r['wall'] for r in commands if r['kind']!='timing'),3))+' s. Go vet overhead is additional and was not separately timed.\n\n'
report+='Not covered: absent frozen corpus, unassigned package rows, package/repo-wide uniqueness, dynamic per-function reachability, and production-entry probes for witness/setup/helper rows. Their vacuous fields remain null. No production mutant edits the Go cohere oracle. Production source is restored before the evidence commit.\n'
(p/'REPORT.md').write_text(report)
(p/'matrix.json').write_text(json.dumps({'base':(p/'base.txt').read_text().strip(),'bounded':True,'matrix_rows':matrix,'production':{id:failed(id) for id in production},'probe':{'P1':failed('P1')},'harness':{e['id']:failed(e['id']) for e in experiments if e['kind'] in ['setup','witness']}},indent=2)+'\n')
print('\n'.join(summary))
