import pathlib,json,subprocess,gzip,hashlib
p=pathlib.Path('review/test-defend/internal-lower-module_namespace');scope=json.loads((p/'scope.json').read_text());audit=json.loads((p/'audit-rows.json').read_text());matrix=json.loads((p/'matrix.json').read_text());bounded=json.loads((p/'D06-bounded-matrix.json').read_text())
assert bounded[0]['rows_failed']==['TestNamespaceCallGraphLinearWork']
assert bounded[1]['exit']!=0
isolated=json.loads((p/'D06-enum-alone.json').read_text())
whole=json.loads((p/'D06-enum-whole.json').read_text())
assert isolated['exit']==0 and whole['exit']==0
byid={r['mutant']:r for r in matrix}; byname={r['test']:r for r in audit};rows=[]
defenses={'TestModuleNamespaceInitializedReadProof':'D01','TestCallableNamespaceReceiverStaysLoud':'D02','TestNamespaceAmbientContextsDoNotExecute':'D03','TestParserFactoryBindingHoisting':'D04','TestNamespaceCallGraphLinearWork':'D06'}
attempt_ids={n:[m] for n,m in defenses.items()};attempt_ids['TestNamespaceEnumInitializationIndependentOfModuleAnalysis']=['D05'];attempt_ids['TestNamespaceCallGraphCycleUnion']=['D07'];attempt_ids['TestNamespaceCallGraphLinearWork']=['D06']
reasons={
 'TestNamespaceCallGraphLinearWork':'D06 fails depth 12 with 8191 walks rather than 13. The same empty diamond graph exists in TestEnumInitializationReach/call-graph-24 and times out there too. One aimed attempt, not three; cannot judge whether a different unique defense exists.',
 'TestNamespaceClassEarlyConstructionStaysLoud':'24 exclusive covered lines identify construction and new edges. No construction-specific mutant was run within the seven-mutant cap; cannot judge. The broad D07 union replacement does not make this row fail.',
 'TestNamespaceAmbientHostInitialization':'16 exclusive covered lines include host and ownership paths. The row has acquired an executable refusal assertion since audit. No host-specific mutant was run within the cap; cannot judge.',
 'TestNamespaceCallGraphCycleUnion':'No exclusive covered lines. Three namespace-only cycle members differ semantically from the two-member enum/namespace subsumer. D07 reproduces a shared union failure, not a unique three-member boundary fault. Fewer than three aimed attempts; cannot judge.',
 'TestNamespaceEnumInitializationIndependentOfModuleAnalysis':'D05 makes this row fail, but TestEnumInitializationReach and TestNamespaceLimitsStayLoud also fail. Fewer than three aimed attempts within cap; cannot judge uniqueness.',
 'TestTscNamespaceDeclarationShapes':'14 exclusive covered lines include optional namespace local initialization and empty erased declaration bodies. No shape-specific mutant was run within the cap; cannot judge.'}
owner={
 'TestNamespaceCallGraphLinearWork':'Name matches exact expansion-count assertions across depths 12 and 24 and repeat queries. Counts detect actual traversal reuse here, but the row does not measure wall time or all forms of graph complexity. It catches D06 promptly at depth 12, unlike the other row that reaches the binary timeout.',
 'TestNamespaceClassEarlyConstructionStaysLoud':'Name matches the NotYet refusal assertion. It does not test native construction behavior, but does not promise to.',
 'TestNamespaceAmbientHostInitialization':'Current row checks both ambient preflight and executable early-read refusal. Host lowering checks only error text, and cwd admits a separate ownership refusal. A different node: error can satisfy the host assertion.',
 'TestNamespaceCallGraphCycleUnion':'Name matches exact namespace identities, full cycle membership and expansion count. No identified name/assertion mismatch.',
 'TestNamespaceEnumInitializationIndependentOfModuleAnalysis':'Name matches the direct namespace preflight call without module scheduling. Its oracle only demands an enum diagnostic substring; it does not separately assert module-analysis state.',
 'TestTscNamespaceDeclarationShapes':'Tsc names fixture provenance, not an executed tsc oracle. Positive cases assert only acceptance, not the lowered contents; overload policy is self-derived.'}
for name in scope['targets']:
 old=byname[name];ids=attempt_ids.get(name,[]);attempts=[];failure=[]
 for ident in ids:
  r=byid[ident];failed=r['rows_failed']
  if ident=='D06': failed=bounded[0]['rows_failed']
  attempts.append({k:r[k] for k in ['mutant','file_line','change']}|{'rows_failed':failed})
  events=bounded[0]['errors'] if ident=='D06' else r['errors']
  failure.extend(e['Output'].strip() for e in events if e.get('Test','').split('/')[0]==name)
 defended=name in defenses;ident=defenses.get(name)
 evidence='; '.join(failure)
 if ids:
  r=byid[ids[0]];evidence='ADAMIC_BUILD_CACHE_DIR='+r['env']['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(r['command'])+'; '+evidence
  if ids[0]=='D06':evidence='ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/D06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestNamespaceCallGraphLinearWork$/^12$; '+ '; '.join(failure)+'; D06-rest.log times out in TestEnumInitializationReach/call-graph-24; D06-enum-alone.log passes its expensive subcase; D06-enum-whole.log passes the complete enum row. Other rows finish passing in D06-rest.log; the combined timeout does not prove an individual catch'
 else:evidence='go test -count=1 -timeout 90s ./internal/lower/ -run ^'+name+'$ -coverpkg=./internal/lower -coverprofile=coverage/'+name+'.out; clean coverage passes; '+reasons[name]
 row={'test':name,'package':'internal/lower','prior_verdict':old['verdict'],'subsumed_by':[] if defended else old['subsumed_by'],'prior_subsumed_by':old['subsumed_by'],'defense':'defended' if defended else 'cannot-judge','unique_mutant':(ident+' '+byid[ident]['file_line']) if ident else None,'attempts':attempts,'evidence':evidence}
 if name=='TestNamespaceEnumInitializationIndependentOfModuleAnalysis':row['subsumed_by']=['TestEnumInitializationReach','TestNamespaceLimitsStayLoud']
 if not defended:row['reason']=reasons[name];row['owner_finding']=owner[name]
 if name=='TestNamespaceCallGraphLinearWork':
  row['bounded']=True;row['bound']='Target depth 12 catches; target depth 24 exceeds 90s. Combined exclusion matrix cooks in enum row, but that complete row separately passes within 90s; every other top-level row passes in exclusion matrix.'
  row['attempts'][0]['combined_run_timeouts']=['TestEnumInitializationReach']
  row['attempts'][0]['isolated_enum_exit']=whole['exit']
  row['subsumed_by']=[]
 rows.append(row)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
unique=[]
for name,ident in defenses.items():
 r=byid[ident];passed=sorted(set(bounded[1]['rows_passed']+['TestEnumInitializationReach'])) if ident=='D06' else r['rows_passed']
 unique.append({'test':name,'mutant':ident,'rows_passed':passed,'skipped_top_level':['TestOriginalCycleLedger','TestOptionalWideningCensus'],'excluded_target_subcase':'24' if ident=='D06' else None})
(p/'unique-pass-lists.json').write_text(json.dumps(unique,indent=2)+'\n')
coverage=json.loads((p/'coverage-runs.json').read_text());vet_seconds=sum(r['vet_seconds'] for r in matrix);mutation_seconds=sum(r['wall_seconds'] for r in matrix);rerun_seconds=sum(r['wall_seconds'] for r in bounded)+isolated['wall_seconds']+whole['wall_seconds']
report=f'''Five rows defended with unique production catches; five remain cannot-judge within the seven-mutant cap.
Current main {scope['base']}; 275 top-level tests, 36 added since audit, all ten target names present.
Tests unchanged; full-package matrices and bounded linear-work recovery, diffs, coverage and logs retained.

'''+json.dumps(rows,indent=2)+f'''

Brief ambiguities, limits and time costs

1. The report is from older main f91994f019703ba25d2918cf529c0e0b0c05d93c. Current main adds 36 tests and strengthens at least three selected rows: initialized reads require executable output and a console write; ambient-host preflight now asserts an executable refusal; parser-factory now checks specific binding metadata and a nonempty body. Old vacuity claims do not transfer unchanged. All selected rows remain in their original files.
2. Ten rows times three attempts would require thirty mutants. The explicit near-minute full-matrix cap allows seven. I prioritized unique behavior and do not claim three honest aimed attempts for the remaining five rows. They remain cannot-judge, not deletion candidates. D07 is a shared-union fallback replay, not a new proof of a three-member-specific fault.
3. Exclusive coverage is a lead. The parser metadata row has zero exclusive lines but asserts metadata others do not inspect. The linear-work row has zero exclusive lines but feeds empty reach sets; enum memoization feeds nonempty sets. TestEnumInitializationReach includes the same empty diamond graph, revealed by a combined-run timeout; its full isolated row passes. It checks acceptance and does not assert body expansion counts. D06 changes actual reuse of empty cached graphs without altering the expansion counter.
4. D06's whole-package binary times out at 90.345 seconds. The timeout panic aborts the binary; do not use that original run for uniqueness. Depth 12 alone fails with 8191 expansions versus 13. The all-other-rows exclusion matrix also times out, identifying TestEnumInitializationReach/call-graph-24 as its only unfinished row. All its other top-level rows finish, except baseline skips. That enum subcase then passes alone in 57.551s, and the complete enum row passes alone in 63.091s too. The combined timeout therefore does not prove an individual row failure. The target depth-12 row fails while every other top-level row has an observed pass across the exclusion matrix and complete enum rerun. This is a bounded unique defense; target depth 24 remains over budget.
5. Defended means unique among the executed current package rows. TestOriginalCycleLedger and TestOptionalWideningCensus require external project inputs and skip throughout. One mixed-union subcase is deferred compiler implementation. These skipped inputs remain unknown; no toolchain installation enables them. Scope and raw logs preserve their names. No repo-wide uniqueness is claimed.
6. Seven standalone diffs use starting-main line positions, apply independently, and pass go vet ./internal/lower/. No test, harness, fixture, external oracle or checker was modified. D05 only changes diagnostic wording; it proves the text oracle fires but does not establish an independent soundness check.
7. The final schema has only defended, not defended and cannot-judge, while the prose introduces subsumed for newly caught formerly untrue rows. No conflict is needed here: the formerly untrue parser row is uniquely defended; the other two remain unattempted within the cap.
8. Automatic approval review initially failed with an internal Habitat service error before dependency installation. The same authorized operation succeeded on retry. No unsafe-action rejection or user approval was needed.
9. Go coverage instrumentation initially compiled for 10.38 seconds; later solo commands took about two seconds including Go overhead. Rest-of-package coverage required three additional full runs. Their exclusion regexes and profiles are saved. Independent source reads and test listing establish the current scope.

Owner findings for unresolved rows and performance scope

'''+ '\n'.join('- '+n+': '+owner[n] for n in scope['targets'] if n in owner and n not in defenses)+f'''

Timing and exclusions

Warm toolchain worked; setup skipped. npm ci reported 630ms; nproc 5; Go1.27.1. Clean baseline binary 41.705s. Coverage commands totaled {sum(r['seconds'] for r in coverage):.3f}s wall. Seven standalone vet checks totaled {vet_seconds:.3f}s. Seven full mutation commands totaled {mutation_seconds:.3f}s wall, including compilation and native builds; bounded reruns totaled {rerun_seconds:.3f}s. Each mutation uses ADAMIC_BUILD_CACHE_DIR=/tmp/defend-module-namespace/cache/Dxx. Separate native rebuild durations are not emitted by these tests; command wall and binary elapsed are recorded, not inferred.

No more than seven production mutants, no full depth-24 result for D06, no mutants aimed specifically at the unresolved host, shape or construction leads, no three-attempt rejection verdicts or claimed uniqueness based only on timeout-aborted runs, no skipped external-input coverage, no other package or repo-wide replay. No tests were deleted, rewritten or weakened. Raw JSON logs may be losslessly gzip-compressed after report generation; commands name the original log destinations.

Evidence files: rows.json, code-and-oracle.md, scope.json, coverage-exclusive.json, coverage-runs.json, coverage/*.out, mutant-plan.json, matrix.json, D06-bounded-matrix.json, unique-pass-lists.json, diffs/*.diff, logs/*, and execution scripts. Production source is restored byte for byte before evidence commit.
'''
(p/'REPORT.md').write_text(report)
print('report ready',len(rows),'rows')
