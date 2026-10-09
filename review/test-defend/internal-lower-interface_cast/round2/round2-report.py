import pathlib,json,subprocess,shutil
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-interface_cast/round2'
def events(f):
 es=[]
 for l in f.read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return es
plans=json.loads((P/'plan.json').read_text());runs=json.loads((P/'runs.json').read_text());matrix={}
for run in runs:
 if run['selector']:
  es=events(P/(run['id']+'.log'));out={e['Test']:e['Action'] for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']}
  matrix[run['id']]={'rows':out,'failed':[n for n,a in out.items() if a=='fail'],'passed':[n for n,a in out.items() if a=='pass'],'skipped':[n for n,a in out.items() if a=='skip'],'binary_seconds':next((e.get('Elapsed') for e in reversed(es) if not e.get('Test') and e.get('Action') in ['pass','fail']),None),'wall_seconds':run['seconds']}
(P/'matrix.json').write_text(json.dumps(matrix,indent=2))
rows=[]
for name,subsumer in [('TestDefaultTaggedInterfaceNeedsNoFlag','TestDefaultTaggedInterfaceAdmission'),('TestIteratorGapsAreExplicit','TestIteratorViewsCannotEraseReceivers')]:
 attempts=[];proof=[]
 for plan in plans:
  id=plan['id']
  if plan['test']!=name or id not in matrix:continue
  attempts.append({'mutant':id,'file_line':plan['file']+':'+str(plan['line']),'change':plan['kind']+': '+plan['old'].strip()+' -> '+(plan['new'].strip() or '(deleted)'),'rows_failed':matrix[id]['failed']})
  for e in events(P/(id+'.log')):
   if e.get('Test','').split('/')[0]==name and '.go:' in e.get('Output',''):proof.append(id+': '+e['Output'].strip())
 unique=next((a for a in attempts if a['rows_failed']==[name]),None)
 rows.append({'test':name,'package':'internal/lower','prior_verdict':'subsumed','subsumed_by':subsumer,'defense':'defended' if unique else 'not defended','unique_mutant':unique['mutant']+' '+unique['file_line'] if unique else None,'attempts':attempts,'evidence':'; '.join(proof),'command':'ADAMIC_MUTANT=<id> ADAMIC_BUILD_CACHE_DIR=/workspace/defend-lower-round2-cache/<id> timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > <id>.log 2>&1'})
(P/'rows.json').write_text(json.dumps(rows,indent=2))
for f in ['round2-matrix.py','round2-coverage.py','round2-report.py']:shutil.copy('/workspace/'+f,P/f)
for f in ['audit.json','matrix.json']:
 (P/'prior'/f).write_bytes(subprocess.check_output(['git','show','origin/test-audit/internal-lower-interface_cast:review/test-audit/internal-lower-interface_cast/'+f],cwd=R))
check=[]
for plan in plans:
 r=subprocess.run(['git','apply','--check',str(P/(plan['id']+'.diff'))],cwd=R,capture_output=True,text=True);check.append({'id':plan['id'],'exit':r.returncode,'output':r.stdout+r.stderr});assert r.returncode==0
(P/'apply-check.json').write_text(json.dumps(check,indent=2))
text='''# Two-row defense, round 2

Starting origin/main: {sha}. nproc: 5. Warm env.sh worked; setup skipped. npm ci stage3/api: 1.201 seconds. Baseline green, wall 68.727 seconds. All 275 discovered top-level tests completed in each matrix. Two rows skipped: TestOriginalCycleLedger and TestOptionalWideningCensus. Uniqueness means the default enabled package suite; those skipped inventories remain unknown. No run exceeded the binary budget, and no panic interrupted discovery.

CODE UNDER TEST: Adamic's Go Lower implementation, especially checked tagged-interface casts/view field access and custom iterator array spread representation validation. ORACLE: NeedsNoFlag executes original TypeScript on Node and compares lowered JavaScript stdout/stderr/exit through the existing agreement helper (external-run). IteratorGaps uses hand-written expectations that errors.As finds *NotYet (self). No oracle or test was changed.

Coverage: NeedsNoFlag has zero exclusive covered source lines versus DefaultTaggedInterfaceAdmission. That table's unused-bad-factory case includes the same inline positive call/payload read, plus an unused function. Three shared-line semantic attempts were made: required-field admission, allowed-tag membership, and discriminant selection. All are caught by the admission row too. This finite defense did not establish uniqueness, and is not a deletion recommendation.

IteratorGaps has 1420 exclusive covered source lines versus IteratorViewsCannotEraseReceivers, across its seven inputs. Its final custom iterator number spread into (number|undefined)[] reaches the outer arrayLiteral guard at object.go:237. Dropping that refusal admits the unsupported representation and makes only IteratorGaps fail at iteration_test.go:30. The previous inner collection guard was not the guard deciding this case. G01/G02 were planned and vet-validated but not run after G03 established the defense. Only executed attempts are in rows.json.

Every standalone diff applies to the starting tree and passed go vet ./internal/lower/. Mutants were selected from a switched production source in one build, each with a separate ADAMIC_BUILD_CACHE_DIR. Switched source was restored. No tests, oracle or harness changed. matrix.json contains every passed, failed and skipped row; raw JSON logs contain subcases. plan.json maps edits to starting-commit lines. coverage-differences.json records exclusive blocks and lines.

Name/assertion finding for the row not defended: NeedsNoFlag checks a positive cast under the inherited environment and Node agreement, but does not explicitly unset or toggle a tagged-interface feature flag. It demonstrates current default behavior rather than proving independence from every flag configuration. IteratorGaps checks only the NotYet category, not each refusal reason; this allowed earlier unrelated refusals to mask disabled inner guards, but G03 removes the actual outer guard and yields nil.

Brief feedback and time costs:
- The requested branch already existed from the previous wider defense. Work began detached at freshly fetched origin/main, then evidence was merged into the existing defense branch, preserving history without a force push.
- The audit's report is split across REPORT.md, rows.json, summary.json and other evidence. All report/code/oracle notes were read; audit.json and matrix.json are also retained.
- The package gained tests: 239 in the prior audit, 275 now. Current NeedsNoFlag assertions also execute Node agreement, stronger than the old audit's nil-error oracle. Historical conclusions cannot be transferred without rechecking bodies.
- Whole-package uniqueness cannot include inventory rows skipped behind optional corpus configuration. Those two are identified rather than reported as passing.
- No exclusive statement lines is not proof of duplication. The identical positive call in the admission table motivated three shared-line semantic mutations, which remain nonunique.
- An ambiguous optional-field anchor occurred twice, including legacyView. It was rejected before edits and narrowed to the intended unique guard; this was an execution correction, not a defect in the brief.
- The 90-second binary budget differs from wall time: G03 took 82.85 wall seconds including compilation; binary durations are in matrix.json. No narrowing was needed.

Timings: standalone vet checks and switch vet in runs.json, coverage runs in coverage-times.json, baseline setup/list timing in baseline-times.json. Matrix total wall seconds: {total:.3f}. Full package runs, including all new enabled tests, were used. No repo-wide suite, native-only opt-in rerun, or skipped inventory corpus was covered. No main push and no PR.
'''.format(sha=(P/'starting-commit.txt').read_text().strip(),total=sum(x['wall_seconds'] for x in matrix.values()))
(P/'REPORT.md').write_text(text)
print(json.dumps(rows,indent=2));print('matrix',[(k,len(v['passed']),len(v['failed']),v['binary_seconds']) for k,v in matrix.items()])
