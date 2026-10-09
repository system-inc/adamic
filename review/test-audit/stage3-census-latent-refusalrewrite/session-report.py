import pathlib,json,re,statistics,csv,datetime
p=pathlib.Path('/workspace/adamic/review/test-audit/stage3-census-latent-refusalrewrite');inv=json.loads((p/'inventory.json').read_text());plan=json.loads((p/'plan.json').read_text());runs=json.loads((p/'runs.json').read_text());rows=inv['rows'];ids=['M'+str(i) for i in range(1,13)];matrix={};details={}
for m in plan:
 events=[]
 for line in (p/(m['id']+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 matrix[m['id']]=[e['Test'] for e in events if e['Action']=='fail' and 'Test' in e]
 details[m['id']]={r:''.join(e.get('Output','') for e in events if e.get('Test')==r) for r in matrix[m['id']]}
 # Standalone diffs reproduce exactly the switched matrix.
 replay=[json.loads(s) for s in (p/(m['id']+'-replay.log')).read_text().splitlines() if s.startswith('{')]
 replayfails=[e['Test'] for e in replay if e['Action']=='fail' and 'Test'in e]
 assert set(replayfails)==set(matrix[m['id']]),(m['id'],replayfails,matrix[m['id']])
results=[]
for r in rows:
 timings=[float(re.search(r'\t([0-9.]+)s', (p/(r+'-'+str(i)+'.log')).read_text()).group(1)) for i in range(1,4)]
 witness=r in rows[1:3];kills=[] if witness else [i for i in ids if r in matrix[i]];unique=[i for i in kills if matrix[i]==[r]]
 last={'TestCompiler41231d51Shape':'M2','TestMissingFunctionMutantFailsLoudly':'W1','TestChangedVisitorMutantFailsLoudly':'W2','TestVisitorsCollectContinueAndSkipDiagnosedBodies':'M9'}[r]
 fail=next(s.strip() for s in details[last][r].splitlines() if 'rewrite_test.go:'in s)
 oracle={rows[0]:'Self-written: production refuse AST must be unchanged and latentRefuse must contain exactly four defers. Counting defers does not check owner behavior.',rows[1]:'Self-written missing-function diagnostic substring and nil partial output. Exact label matters: M12 fails although rejection still works. Witness of Rewrite validation.',rows[2]:'Self-written changed-child-walk diagnostic substring. Exact label matters: M12 fails although rejection still works. Witness of Rewrite validation.',rows[3]:'Self-written ordered findings, continuation, diagnosed-body exclusion and first production refusal, executed in a synthetic Go program. Child go test must exit zero; logs distinguish compilation errors from semantic mismatches.'}[r]
 result=dict(test=r,package='stage3/census/latent/refusalrewrite',file='stage3/census/latent/refusalrewrite/rewrite_test.go',seconds=statistics.median(timings),oracle=oracle,oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=last+': '+fail,verdict='witness' if witness else 'sacred',subsumed_by=[],mutants_in_matrix=ids,probe_kills=['P1'] if r in matrix['P1'] else [],subsumer_seconds=None,vacuous=r not in matrix['P1'],bounded=False,matrix_rows=rows,evidence='ADAMIC_MUTANT='+last+' timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/refusalrewrite/ -run . > '+last+'.log 2>&1; '+fail,timing_samples=timings)
 if witness:result['witness_kills']=[last]
 results.append(result)
(p/'results.json').write_text(json.dumps(results,indent=2)+'\n')
with (p/'matrix.csv').open('w') as f:
 writer=csv.writer(f);writer.writerow(['id','kind',*rows]);
 for m in plan:writer.writerow([m['id'],'probe' if m['id'].startswith('P') else 'witness check' if m['id'].startswith('W') else 'production',*['fail' if r in matrix[m['id']] else 'pass' for r in rows]])
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
initial=json.loads((p/'initial.json').read_text());costs={'setup_seconds':0,'npm_ci_seconds':initial['npm']['seconds'],'clean_baseline_driver_seconds':initial['baseline']['seconds'],'clean_baseline_binary_seconds':0.237,'isolated_12_runs_driver_seconds':sum(v['wall_seconds'] for k,v in runs.items() if k.startswith('Test')),'switch_build_seconds':runs['switch-build']['wall_seconds'],'clean_build_seconds':runs['clean-build']['wall_seconds'],'final_12_mutant_matrix_driver_seconds':sum(runs[i]['wall_seconds'] for i in ids),'witness_matrix_driver_seconds':sum(runs[i]['wall_seconds'] for i in ['W1','W2']),'probe_matrix_driver_seconds':runs['P1']['wall_seconds'],'standalone_replay_driver_seconds':sum(v['wall_seconds'] for k,v in runs.items() if k.endswith('-replay')),'final_vet_driver_seconds':sum(v['wall_seconds'] for k,v in runs.items() if k.endswith('-vet')),'note':'Driver timings include Go tool startup and any compilation; recorded final runs exclude overwritten development retries. All steps below 90 seconds. No native rebuild needed.'};(p/'costs.json').write_text(json.dumps(costs,indent=2)+'\n')
summary='''u160 audited four rows at cf735d9fba9e38de6368575e5630e44375a86eaf.
Clean baseline passed; no skips; warm setup skipped; nproc=5.
Twelve production mutants, two weakened checks and one empty-entry probe ran over the whole package.
Two rows are sacred; two are witnesses; no row passed its own empty-entry probe.
M10 and M11 survived with observed output differences; repo-wide uniqueness remains untested.
'''
table='| ID | File:line at starting origin/main | Change | Failed rows |\n|---|---|---|---|\n'
short={rows[0]:'Shape',rows[1]:'Missing',rows[2]:'Changed',rows[3]:'Behavior'}
for m in plan:
 description=m['menu']+': '+m['old'].strip().replace('\n',' ')+' -> '+(m['new'].strip().replace('\n',' ') or '(drop)')
 if 'compile_adjustment'in m:description+='; '+m['compile_adjustment']
 table+='| '+m['id']+' | rewrite.go:'+str(m['line'])+' | '+description.replace('|','\\|')+' | '+', '.join(short[r] for r in matrix[m['id']])+' |\n'
text=summary+'\n```json\n'+json.dumps(results,indent=2)+'\n```\n\n'+table+'''
All source locations above refer to stage3/census/latent/refusalrewrite/rewrite.go at the stated starting commit. Shape, Missing, Changed and Behavior abbreviate the four rows in JSON order. M IDs are production mutants; W IDs only support witness verdicts; P1 is only a vacuity probe. No families, helpers, setup checks, panics of the parent binary, timed-out runs or skipped rows occurred. M5 panicked in the child behavior process, whose captured failure did not abort the package matrix. Every standalone diff passed git apply --check and go vet, and its whole-package replay matched the switched matrix.

Survivors:
- M10: go run ./stage3/census/latent/refusalrewrite/cmd -input stage3/census/latent/refusalrewrite/testdata/refusals-41231d51.go.txt -output <artifact> wrote `if latentFullEnabled() && node.Kind == ast.KindFunctionDeclaration {` before and `if latentFullEnabled() || node.Kind == ast.KindFunctionDeclaration {` after. This is unguarded generated control flow. The witness demonstrates generated output, not a runtime full-mode behavior comparison. See M10-witness-before.go.txt and M10-witness-after.go.txt.
- M11: the same command with two-outer-returns.go.txt succeeded before (exit 0, generated output saved) and failed after (exit 1, `latent refusal rewrite: lowering.refuse: unsupported outer refusal returns`). Unguarded acceptance of a supported two-return input. See M11-witness-before.go.txt and M11-witness-after.log.

Brief ambiguities and costs:
- The brief names no row list for a whole package, correctly resolved by go test -list. Four distinct bodies, no family merging was appropriate.
- Tests named MutantFailsLoudly validate production Rewrite input guards, rather than an external agreement comparator. Following the explicit witness rule, W1 returns the unchanged input when the missing-function guard fires; W2 permits zero child walks. Their failures alone support witness verdicts. Production diagnostic mutation M12 is recorded in the raw matrix but excluded from their verdict kills.
- The behavior test runs Go, but the brief's external-run category names Go cohere, not arbitrary Go execution. Its ordered expected findings are self-written, so all four rows are self.
- No external source is cited by these expected values; none was checked against an outside authority.
- The shape test's defer count catches M2, but does not prove owner restoration semantics. The behavior fixture pins latentFullEnabled to false and does not exercise full-mode restoration.
- A child nonzero exit alone is weaker than a diagnostic or value comparison. M7, M8 and M9 fail by compilation; M3 and M4 additionally demonstrate semantic findings mismatches. The complete child output is retained in each log.
- Drop-statement mutants required cleanup of dead bindings or fmt arguments. M2 also drops the unused capture; M3 removes the corresponding fmt argument; M8 and M9 discard unused AST bindings. These are compilation repairs, not inserted behavioral statements.
- Initial M8 and P1 replay diffs failed vet for an unused binding and unreachable code. Corrected them, reran vet, and independently replayed all final diffs. Final P1 replaces the entire Rewrite body with return nil,nil. The switch uses an entry return guarded by its selector.
- At most three mutants per row conflicts mildly with witness-only verdict rules. Twelve production mutants were selected across generation, traversal, bounds and diagnostics before any matrix failures were inspected; two separate guard weakenings test witnesses.
- A runtime guard survivor can be witnessed by its generated source output because Rewrite's output is Go source. M10's runtime effect itself was not measured.

CODE UNDER TEST: Rewrite and package helpers fail, ident, printed, statements, eraseWalks, collectReturns, rewriteBlocks. This is the static package call graph reached by these rows, including Rewrite's AST callbacks; no dynamic coverage profile was produced. cmd/main.go is only used to observe survivor outputs and was never mutated. ORACLE: the test's self-written AST counts, diagnostic substrings and ordered behavior findings, all retained unchanged.

Timing and scope: warm env verification succeeded; setup 0 seconds; stage3/api npm ci completed before baseline. See costs.json for separated build and recorded run driver times, runs.json for every final command cost, and the twelve isolated logs for binary timing samples. Native products were not involved. No other package's tests ran. Package uniqueness is proven only for the observed production matrix, not repo-wide. The source was restored and a clean post-audit baseline passed. No push to main and no pull request.
'''
(p/'README.md').write_text(text)
print(json.dumps(costs,indent=2));print('report written')
