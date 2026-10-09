import pathlib,json,subprocess,shutil
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-interface_cast';plans=json.loads((P/'plan.json').read_text());matrix={}
for x in plans:
 es=[]
 for line in (P/(x['id']+'.log')).read_text().splitlines():
  try:es.append(json.loads(line))
  except:pass
 done=[e for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']]
 matrix[x['id']]=dict(rows_failed=[e['Test'] for e in done if e['Action']=='fail'],rows_passed=[e['Test'] for e in done if e['Action']=='pass'],rows_skipped=[e['Test'] for e in done if e['Action']=='skip'],completed_rows=len(done),binary_seconds=[e.get('Elapsed') for e in es if not e.get('Test') and e.get('Action') in ['pass','fail']],output=[e['Output'].strip() for e in es if e.get('Output') and '_test.go:' in e['Output']])
 assert len(done)==275,(x['id'],len(done))
(P/'matrix.json').write_text(json.dumps(matrix,indent=2))
codes=json.loads((P/'code-and-oracles.json').read_text());coverage={r['test']:r for r in json.loads((P/'coverage-differences.json').read_text())};report=[]
limitations={
'TestDefaultTaggedInterfaceNeedsNoFlag':('No exclusive coverage. Its source duplicates an admitted inline-cast use within the subsumer, except for an unused factory declaration. The seven-run budget was spent on distinct safeguards; no aimed mutant for this row.','The test executes Node agreement now, but does not unset or vary an admission flag. The name promises flag independence beyond an assertion under the inherited environment.'),
'TestLiteralMethodViewsDoNotLoseThis':('The row now checks Node agreement. Its positive path is distinct from the refusal subsumer, but is also exercised by the new TestLiteralMethodSignatureViewsDoNotLoseThis. No aimed receiver-behavior mutant fit the seven-run budget.','No observed mismatch between the name and current JavaScript receiver assertion: Node agreement observes this.value. Native execution is not checked, but the name does not explicitly promise native execution.'),
'TestGenericIteratorViewsPreserveNativeArguments':('The row reaches an earlier nominal refusal, with no lowered product. Coverage and source were inspected, but no aimed nominal-argument mutant fit the seven-run budget.','Yes. The name promises native argument preservation; assertions check only Refused plus nominal ancestry in What, before native arguments exist.')}
for c in codes:
 name=c['test'];aimed=[x for x in plans if x['test']==name];unique=[x for x in plans if matrix[x['id']]['rows_failed']==[name]];observed=[x['id'] for x in plans if name in matrix[x['id']]['rows_failed']]
 attempts=[dict(mutant=x['id'],file_line=x['file']+':'+str(x['line']),change=x['old'].strip()+' -> '+(x['new'].strip() or '(drop statement)'),rows_failed=matrix[x['id']]['rows_failed']) for x in aimed]
 if unique:
  x=unique[0];defense='defended';why='Only this row failed in the executable default package matrix; two inventory rows skipped. '+('This defense protects diagnostic wording, not refusal category or runtime semantics.' if x['id'] in ['D01','D02','D03','D04'] else 'The mutant changes the guarded production behavior.');id=x['id']
 else:
  defense='cannot-judge';why=limitations.get(name,('One aimed attempt survived: no row failed. Three honest aimed attempts were not completed within the seven-run budget.',''))[0];id=aimed[0]['id'] if aimed else None
 output=next((s for s in matrix[id]['output'] if ('iteration_test.go:' in s or 'interface_cast_test.go:' in s)),None) if id else None
 command=('ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR=/workspace/defend-lower-cache/'+id+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > '+id+'.log 2>&1') if id else 'See coverage/'+name+'.log and coverage/'+name+'-difference.txt; no aimed mutant run'
 mismatch=limitations.get(name,('', 'Assertions name the rejection category and requested diagnostic substring; no additional name/assertion mismatch established.'))[1]
 if name=='TestIteratorGapsAreExplicit':mismatch='Checks only NotYet category; does not establish which iterator gap caused rejection. The name promises explicit gaps more broadly than this category assertion.'
 if name=='TestDestructuredMethodsCannotLoadOwnSlots':mismatch='Accepts either Refused or NotYet without a reason. An unrelated refusal can satisfy it, so the asserted cause is weaker than the named own-slot safeguard.'
 report.append(dict(test=name,package='internal/lower',prior_verdict='subsumed',subsumed_by=c['subsumer'],defense=defense,unique_mutant=unique[0]['id']+' '+unique[0]['file']+':'+str(unique[0]['line']) if unique else None,attempts=attempts,evidence=command+('; '+output if output else ''),code_under_test=c['code_under_test'],oracle=c['oracle'],coverage_only_lines=coverage[name]['row_only_lines'],observed_kills=observed,reason=why,name_assertion_finding=mismatch))
(P/'report.json').write_text(json.dumps(report,indent=2));checks=[]
for x in plans:
 r=subprocess.run(['git','apply','--check',str(P/(x['id']+'.diff'))],cwd=R,capture_output=True,text=True);checks.append(dict(mutant=x['id'],exit=r.returncode,output=r.stdout+r.stderr));assert r.returncode==0
(P/'diff-apply-checks.json').write_text(json.dumps(checks,indent=2));shutil.copy('/workspace/defend-lower-matrix.py',P/'matrix-runner.py');shutil.copy('/workspace/defend-lower-coverage.py',P/'coverage-runner.py');shutil.copy('/workspace/defend-lower-report.py',P/'report-runner.py')
base=(P/'starting-commit.txt').read_text().strip();runs=json.loads((P/'runs.json').read_text());times=json.loads((P/'baseline-times.json').read_text());covtimes=json.loads((P/'coverage-times.json').read_text())
text='Defense evidence at origin/main '+base+'.\n\n'+json.dumps([{k:r[k] for k in ['test','defense','unique_mutant','reason','name_assertion_finding']} for r in report],indent=2)+'\n\n'
text+='''CODE UNDER TEST and ORACLE: code-and-oracles.json and report.json. Production only was mutated. No test, harness, Node oracle, or preparation code was changed.
Coverage: individual row and named subsumer profiles instrument internal/lower. coverage-differences.json lists covered blocks and source lines exclusive to each row. profile paths refer to the starting commit. NeedsNoFlag has no exclusive lines; shared-line semantic input differences were also inspected.
Matrix: all 275 discovered top-level rows have observed pass/fail/skip results under every mutant. matrix.json contains exact passing lists for each defended row, failing lists and raw failure output. Native product caches are separate per mutant.
Standalone diffs: D01.diff through D07.diff; all passed go vet ./internal/lower/ and apply checks against the restored starting tree. Selector instrumentation was removed; production sources and tests are byte-identical to the starting commit.
Four mutants change production diagnostic constants. They defend only wording contracts, not rejection semantics. D05 drops literal-method closure registration; D06 allows the callee exception for a destructured method view; D07 disables the collected-element representation guard.

Brief feedback, costs and limits:
* The audit report is named REPORT.md, rather than report.md. Initial lookup failed, then the branch tree revealed its name. I initially read the test files during discovery before correcting the report lookup; the report was read before mutation planning or runs.
* Main changed from the audit's 8171b317 to the starting commit above. Package discovery grew from 239 to 275 tests. Positive admission/receiver tests were strengthened to Node agreement, and a signature-view receiver row was added. Old vacuity findings do not describe those current tests.
* Exclusive coverage is not a verdict. Generic refusal and positive receiver rows have exclusive lines, but a distinct mutation can still be shared by other rows. NeedsNoFlag has none and duplicates the subsumer's inline-cast behavior with one unused factory absent.
* Ten rows with up to three aimed attempts can require 30 full matrices, while a near-minute package run permits seven. This run obeyed the seven cap. Unattempted or singly attempted nonunique rows are cannot-judge, not not defended. Three failed attempts were not claimed.
* The definition of defended says no other package row catches the mutant, but two project-specific inventory rows skip by default: TestOriginalCycleLedger and TestOptionalWideningCensus. Their external/project inputs were not prepared. Defenses concern the default executable package gate; behavior under those opt-in inventories is unknown.
* Diagnostic-constant mutants are on the allowed menu and change real production output. Their unique catches prove wording sensitivity only. They do not prove that no other row guards the underlying unsafe admission.
* Whole-file combined reads were sometimes truncated by tool output; focused reads of the relevant test bodies and production safeguards resolved the mutation sites.
* Per-row coverage requires an instrumentation rebuild. The first two coverage commands took about 26s each including compilation, while later commands took about 3s. Commands were run with two workers; their wall durations cannot be summed as elapsed wall time.
* D07 survived the complete executable package matrix. Its before/after Lower witness is saved in D07-witness-clean.log and D07-witness-D07.log. Both runs produced the same earlier spread-representation NotYet. A direct Array.from probe also produced the same earlier optional-widening Refused (D07-witness-from-*.log). No output change was demonstrated, so D07 is an equivalent candidate, not proven unguarded behavior. The witness source and timings are preserved.
* No test was deleted, rewritten or weakened. No repository-wide uniqueness is claimed. No packages outside internal/lower were tested.
* Nondefended name/assertion findings appear per row in report.json. GenericIteratorViewsPreserveNativeArguments only asserts an earlier nominal refusal. NeedsNoFlag does not vary a flag. LiteralMethodViewsDoNotLoseThis now observes receiver behavior through Node, so no name/assertion mismatch was established for its current JavaScript contract.

Timing: warm env.sh worked, setup skipped; nproc=5; npm ci stage3/api '''+str(round(times[0]['seconds'],3))+'''s. Clean package binary 70.193s, command wall '''+str(round(times[-1]['seconds'],3))+'''s. Seven matrix command wall times total '''+str(round(sum(r['seconds'] for r in runs if r['id'] in matrix),3))+'''s. Individual coverage and vet/build timings are in coverage-times.json and runs.json. No production run exceeded the binary's 90s budget.\n'''
(P/'README.md').write_text(text)
print(json.dumps([{k:r[k] for k in ['test','defense','unique_mutant','evidence']} for r in report],indent=2));print(json.dumps({k:v['rows_failed'] for k,v in matrix.items()},indent=2))
