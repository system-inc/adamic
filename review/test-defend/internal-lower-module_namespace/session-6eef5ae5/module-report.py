import pathlib,json,subprocess,shutil
P=pathlib.Path('/workspace/adamic/review/test-defend/internal-lower-module_namespace/session-6eef5ae5');R=P.parents[2];plans=json.loads((P/'plan.json').read_text());matrix=json.loads((P/'matrix.json').read_text());runs=json.loads((P/'runs.json').read_text())
C='TestNamespaceClassEarlyConstructionStaysLoud';A='TestNamespaceAmbientHostInitialization';S='TestTscNamespaceDeclarationShapes';rows=[]
for name in [C,A,S]:
 ats=[];fail_lines=[]
 for x in plans:
  if x['id'] not in matrix or x['test']!=name:continue
  id=x['id'];ats.append(dict(mutant=id,file_line=x['file']+':'+str(x['line']),change=x['kind']+': '+x['old']+' -> '+x['new'],rows_failed=matrix[id]['failed']))
  for l in (P/(id+'.log')).read_text().splitlines():
   try:e=json.loads(l)
   except:continue
   if e.get('Test','').split('/')[0]==name and '.go:' in e.get('Output','') and e.get('Action')=='output' and any(w in e['Output'] for w in ['early constructor read trusted','guard lost','reopened namespace','runtime value','got ','namespace declaration']):fail_lines.append(id+': '+e['Output'].strip())
 unique=next((a for a in ats if a['rows_failed']==[name]),None)
 rows.append(dict(test=name,package='internal/lower',prior_verdict='subsumed' if name==C else 'untrue',subsumed_by='TestCallableNamespaceLimitsStayLoud' if name==C else None,defense='defended' if unique else ('not defended' if len(ats)==3 else 'cannot-judge'),unique_mutant=unique['mutant']+' '+unique['file_line'] if unique else None,attempts=ats,evidence='; '.join(fail_lines),command='ADAMIC_BUILD_CACHE_DIR=/workspace/defend-module-cache/<id> timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > <id>.log 2>&1 (standalone <id>.diff applied)'))
(P/'rows.json').write_text(json.dumps(rows,indent=2));checks=[]
for x in plans:
 if x['id'] not in matrix:(P/(x['id']+'.diff')).unlink(missing_ok=True)
for x in plans:
 if x['id'] not in matrix:continue
 r=subprocess.run(['git','apply','--check',str(P/(x['id']+'.diff'))],cwd='/workspace/adamic',capture_output=True,text=True);checks.append(dict(id=x['id'],exit=r.returncode,output=r.stderr));assert r.returncode==0
(P/'apply-checks.json').write_text(json.dumps(checks,indent=2))
for f in ['module-defense.py','module-coverage.py','module-rest-coverage.py','module-report.py']:shutil.copy('/workspace/'+f,P/f)
base=[]
for l in (P/'baseline.log').read_text().splitlines():
 try:base.append(json.loads(l))
 except:pass
skips=[e['Test'] for e in base if e.get('Test') and '/' not in e['Test'] and e.get('Action')=='skip'];baseline=next(e['Elapsed'] for e in base if not e.get('Test') and e.get('Action')=='pass')
report=f'''# Namespace defense

Starting origin/main: {(P/'starting-commit.txt').read_text().strip()}. nproc: 5. Warm /workspace/adamic-tools/env.sh worked, so setup skipped. npm ci stage3/api completed successfully before the green full-package baseline ({baseline} binary seconds). All 275 current top-level tests completed in each matrix. Default skipped rows: {skips}. Their effects remain unknown; the word unique refers to enabled tests, never those skipped inventories.

CODE UNDER TEST: Adamic's lowering namespaceInitialization, namespaceCallGraph.discover constructor-expression traversal and namespaceRefusal type/namespace merge admission. ORACLE: self-written NotYet category and text assertions in class/ambient rows, and nil-error/NotYet expectations in normalized TypeScript declaration shapes. The shape row derives overload policy from Adamic itself; it does not execute tsc. No oracle, harness or test was mutated.

Coverage profiles precede all mutations. Class has 24 exclusive lines versus CallableNamespaceLimitsStayLoud, recorded in class-exclusive.json. The key semantic difference is a constructor expression nested in a reachable factory instead of a qualified function-namespace property read. C1 adds NewExpression to an existing early-return condition in the production graph walk, bypassing construction discovery. It causes nil instead of the expected NotYet, solely in the class row. This is a condition mutation, not an inserted statement.

Ambient's source changed since the audit: it now first requires executable direct-read preflight to fail. A1 changes the direct-read diagnostic constant, leaving the reachable-call diagnostic unchanged. The row uniquely requires the full 'before runtime initialization' wording on this path. This defense establishes the diagnostic contract of that added assertion, not correct native host behavior. The realpath case requires a node: host diagnostic; cwd may still satisfy its expectation with an independent ownership refusal.

Shapes exercises normalized declaration admissions, including BuilderState and BinaryExpressionState type/value name coexistence. S1 removes the interface exception from namespaceRefusal and affects no observed package result; it is an equivalent candidate for these inputs, not an unguarded-behavior finding. S2 removes the type-alias exception, which wrongly refuses the BinaryExpressionState merged alias/namespace. Only the shape row fails. This establishes admission coverage that the prior fixed menu did not touch. Positive shape assertions require only no error and ignore emitted behavior; admission coverage is not a runtime proof.

Coverage for formerly untrue rows is compared against all other enabled package tests using rest.cover plus the other target's individual profile; untrue-exclusive.json records the differences. Shared lines may still have different inputs, as the distinct merged interface and direct-read diagnostic demonstrate. The rest coverage command and full row list are in rest-coverage-run.json.

Evidence: each executed standalone diff applies to starting origin/main and passed go vet ./internal/lower/. Separate Go source builds were used for these standalone patches; each package matrix used its own ADAMIC_BUILD_CACHE_DIR, avoiding native product reuse. runs.json records binary and command timings, including compilation. matrix.json contains all passed, failed and skipped names. The whole package was run, including all newly added enabled tests. No panic or timeout occurred. Unexecuted alternatives in plan.json were not needed after each defense and support no verdict.

Brief feedback and costs:
- Historical audit evidence was spread across REPORT.txt, report.json, rows.json, mutant-plan.json and matrix.json, rather than the suggested report.md names. These were retained and read.
- The ambient row changed since the audit. Its new executable preflight assertion opens a defense absent from the old evidence; auditing the old body would have missed it.
- The default suite still skips two inventory rows. Absolute package uniqueness cannot be established for them without the external corpus inputs.
- The class matrix exceeded 90 seconds of wall time but remained below 90 seconds in its test binary. This distinction matters for the brief's compile allowance.
- The Tsc name describes fixture provenance; no tsc external-run oracle exists. Its success assertions test admission rather than emitted behavior.
- A diagnostic-constant mutant is explicitly permitted, but its unique failure proves wording, not the underlying ambient-host semantics. This limitation is stated instead of promoting the result into a stronger claim.

No test was deleted, rewritten or weakened. Production source was restored before coverage comparisons and publication. No other package or repo-wide suite was run. No push to main or PR.

Matrix wall total: {sum(x['seconds'] for x in runs):.3f} seconds; binary total: {sum(x['binary_seconds'] for x in runs):.3f} seconds. Coverage and validation commands/timing are saved alongside logs.
'''
(P/'REPORT.md').write_text(report);print(json.dumps(rows,indent=2));print('skips',skips)
