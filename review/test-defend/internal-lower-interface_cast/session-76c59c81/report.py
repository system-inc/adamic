import pathlib,json,subprocess,shlex
p=pathlib.Path('review/test-defend/internal-lower-interface_cast')
meta={i:json.loads((p/(i+'.results.json')).read_text()) for i in ['D01','D02','D03','D04','D05']}
rows=[]
for test,sub,ids,unique in [('TestLiteralMethodViewsDoNotLoseThis','TestDestructuredMethodsCannotLoadOwnSlots',['D01','D03'],'D03'),('TestGenericIteratorViewsPreserveNativeArguments','TestUncheckableCastsStayRefused',['D02','D04','D05'],'D04')]:
 evidence=[]
 for line in (p/(unique+'.log')).read_text().splitlines():
  try:x=json.loads(line)
  except:continue
  if x.get('Test')==test and 'iteration_test.go:' in x.get('Output',''):evidence.append(x['Output'].strip())
 cmd=f'ADAMIC_BUILD_CACHE_DIR=/tmp/defend-interface-cast/cache/{unique} timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > {unique}.log 2>&1'
 rows.append(dict(test=test,package='internal/lower',prior_verdict='subsumed',subsumed_by=[sub],defense='defended',unique_mutant=unique+' '+meta[unique]['file_line'],attempts=[{k:meta[i][k] for k in ['file_line','change','rows_failed']}|{'mutant':i} for i in ids],evidence=cmd+'; '+evidence[0]))
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
mat={}
for ident in meta:
 events=[]
 for line in (p/(ident+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 mat[ident]={x['Test']:x['Action'] for x in events if x.get('Test') and '/' not in x['Test'] and x['Action'] in ['pass','fail','skip']}
(p/'matrix.json').write_text(json.dumps(mat,indent=2)+'\n')
checks=[]
for ident in meta:
 r=subprocess.run(['git','apply','--check',str(p/(ident+'.diff'))],capture_output=True,text=True);checks.append({'mutant':ident,'exit':r.returncode,'output':r.stdout+r.stderr});assert r.returncode==0
(p/'apply-checks.json').write_text(json.dumps(checks,indent=2))
report='''Both requested rows are defended by package-unique production mutations on main 76c59c81e8617cea1892a01841494895927a712c.
Baseline passed in 43.095 binary seconds; the current inventory has 275 top-level tests, versus 239 in the audit.
Five standalone diffs passed go vet and whole-package matrices; source code is restored.

CODE UNDER TEST: Adamic lowering, especially erasedMethodField, callClosure, classViewRefusal, nominalAncestor and sameClassArguments. ORACLE: TestLiteralMethodViewsDoNotLoseThis now runs source and lowered JavaScript on Node and compares stdout, exit and stderr. TestGenericIteratorViewsPreserveNativeArguments checks self-written Refused and the nominal ancestry substring. Neither target executes native output.

Prior evidence was fetched using the full remote refspec, and REPORT.md, rows.json, reached-function inventory and M07/M11 were read. The audit predates stronger agreement assertions and the split TestLiteralMethodSignatureViewsDoNotLoseThis. scope-change.json lists all additions and removals. Every current default row is included in each matrix; matrix.json preserves pass/fail/skip and each results file lists every passing row explicitly.

Coverage commands: timeout 120 go test -count=1 -timeout 90s ./internal/lower/ -run '^NAME$' -coverpkg=./internal/lower -coverprofile=review/test-defend/internal-lower-interface_cast/NAME.cover > NAME.coverage.log 2>&1, for each target and named subsumer. All four pass. Exclusive block spans and the lines inferred from those spans are saved separately. Coverage is statement-block granularity, not proof every character on a listed line executed.

Literal defense: 374 exclusive blocks versus the old subsumer include iteration_origin.go:81 and expression.go:1314. The target accepts a literal method through a function-valued property and executes this.value. Its old subsumer rejects extraction/class-view inputs before call lowering. D01 disables the call exemption, preserving field-read refusal; the target and its newly split method-signature sibling both fail. D03 changes the dispatch option from true to whether the checked property symbol is a method. A function-valued property has no method flag even though the actual object member needs a receiver. Only the target fails: JavaScript stdout is empty instead of Node's 1 newline. The method-signature sibling and old subsumer both pass. D03 is a real production option change, not a test/oracle mutation.

Generic defense: 162 exclusive blocks include contextual classViewRefusal's diagnostic return at class_inheritance.go:575. The subsumer uses explicit assertions, while this row assigns new Sequence<number> to Sequence<number|undefined>. Both cover sameClassArguments:791, but the target supplies a union argument and the subsumer's incompatible-generic case supplies a string widening. D02 bypasses nominal checking for constructor-expression assignments; the target and TestInheritanceRefusesFalseNominalViews fail. D04 loosens the generic identity condition for union targets. Only the target fails because it reaches the later generic-iterable NotYet instead of the required earlier nominal Refused. D05 returns early for iterable constructor assignments and also uniquely fails the target. These aim at assignment position, union representation, and iterable constructor semantics, not test names or source spelling.

All mutant commands use ADAMIC_BUILD_CACHE_DIR=/tmp/defend-interface-cast/cache/ID timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > ID.log 2>&1. Each diff was vetted while applied, then restored. apply-checks.json confirms each applies cleanly against starting main. No tests, harnesses or oracle implementations changed. No matrix panic or timeout occurred. No survivor occurred. No other package was tested. Whole-package default uniqueness is observed, not repository-wide uniqueness.

Limits and brief costs: origin/main advanced since the audit, and the literal row's assertion strength changed materially. Coverage exclusivity only identifies leads; the successful generic defense relies on a semantic difference on shared lines. The fixed menu allows aimed condition/dispatch-option changes; the report describes their domain rather than portraying them as random audit mutants. Two top-level inventories still skip: TestOriginalCycleLedger needs an external pristine compiler diagnostic ledger; TestOptionalWideningCensus needs a caller-selected project configuration. The nested mixed-union interface case also skips in baseline and matrices. These skipped inputs remain unknown, so uniqueness is among the default executable package rows. The user provided no installable ledger/project inputs. Both targets and all named subsumers ran. A local attempt to save coverage differences without filesystem escalation failed before writing; it was rerun with authorization. No results were lost.

Owner finding: the generic row's name suggests preserved native arguments, but its assertions prove only the earlier nominal refusal category/text. No native argument or ABI behavior is executed. This narrower diagnostic contract is independently defended. The literal row now checks actual receiver behavior through lowered JavaScript on Node; its name is supported by that assertion.

Toolchain setup skipped; env.sh works, nproc=5. npm ci succeeded before baseline. Per-matrix vet plus command wall seconds are recorded in ID.results.json. Native cache roots are isolated as required; no standalone native product build was needed for these targets. Production sources are restored and only defense evidence is committed. No main push or pull request.
'''
(p/'REPORT.md').write_text(report+'\n```json\n'+json.dumps(rows,indent=2)+'\n```\n')
print(json.dumps(rows,indent=2));print('matrix wall seconds',{i:round(a['wall_seconds'],3) for i,a in meta.items()})
