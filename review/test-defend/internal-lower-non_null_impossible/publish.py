import pathlib,json,gzip,shutil,subprocess
r=pathlib.Path('/tmp/defend-nonnull');out=pathlib.Path('review/test-defend/internal-lower-non_null_impossible');out.mkdir(parents=True,exist_ok=True);m=json.loads((r/'matrix.json').read_text());assert len(m)==7
for x in m:
 if x['mutant']=='N2':x['file_line']='internal/lower/non_null.go:83';x['additional_file_lines']=['internal/lower/non_null.go:92']
(r/'matrix.json').write_text(json.dumps(m,indent=2)+'\n')
rows=[]
for test,prior,sub in [('TestAdamicNullishAssertionsAreRefused','subsumed','TestNonNullAssertionOnPresentTypeIsErased'),('TestNonNullAssertionOnPresentTypeIsErased','subsumed','TestNonNullAssertionLowersToNullishPanic'),('TestOptionalWideningCensus','untrue',None)]:
 attempts=[x for x in m if x['target']==test];unique=[x for x in attempts if x['rows_failed']==[test]];defense='defended'if unique else'not defended'
 if not unique:assert len(attempts)==3
 chosen=unique[0]if unique else attempts[-1];fail=next((e['Output'].strip()for e in chosen['failures']if e.get('Test','').split('/')[0]==test),None)
 if fail is None:fail='TestOptionalWideningCensus passed; output has files=1 sites=0 (clean files=1 sites=1)'
 row={'test':test,'package':'internal/lower','prior_verdict':prior,'subsumed_by':[sub]if sub else[],'defense':defense,'unique_mutant':chosen['mutant']+' '+chosen['file_line']if unique else None,'attempts':[{k:x[k]for k in ['mutant','file_line','change','rows_failed']}for x in attempts],'evidence':chosen['command']+' > '+chosen['mutant']+'.log 2>&1; '+fail};rows.append(row)
(out/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
for f in r.iterdir():
 if f.is_file()and f.name!='N2-uncompilable.diff':
  if f.suffix=='.log':(out/(f.name+'.gz')).write_bytes(gzip.compress(f.read_bytes()))
  else:shutil.copyfile(f,out/f.name)
shutil.copytree(r/'census-project',out/'census-project',dirs_exist_ok=True)
scope=json.loads((r/'scope.json').read_text());base=scope['base'];totals=[(x['mutant'],len(x['rows_passed']),x['rows_failed'],round(x['wall_seconds'],3))for x in m]
report=f'''One requested row is defended; two are not defended after three production attempts each.
Current origin/main base: {base}; all {len(scope['rows'])} current top-level tests included.
Seven full-package matrices, including the enabled census; no tests rewritten or deleted.

CODE UNDER TEST: Adamic's refusal pass for .a non-null assertions, nonNullValue's present-value erasure paths, and optionalAtSite/optionalValue/optionalWidened production relation analysis. ORACLES: handwritten Refused class, What/Fix strings, NumberConstant/no-panic-string IR checks, and census enumeration/encoding success. None of the three requested rows compares with an external oracle. Census checks neither count nor contents.

Evidence and verdicts

TestNonNullAssertionOnPresentTypeIsErased is defended by N2. The target covers non_null.go:82-84 and recordNonNullCheck's proven branch while its prior panic-check subsumer does not. Its input is present literal 0!, while the subsumer's map.get('a')! is nullable. N2 drops the early returns at origin lines 83 and 92. Both are necessary because the second fast path would otherwise still erase the literal assertion. This is one composite fast-path-drop mutant using two existing return statements, with no inserted statements. The standalone 0! control prints 0 both before and after using node oracle/node.mjs on generated JavaScript, while N2 emits a redundant Coalesce and panic text. The ordinary package agreement rows also passed. Only this requested row failed; all 274 other runnable top-level rows passed, including the prior subsumer. Complete passed lists are in matrix.json. This also guards erasure cost by rejecting redundant checking machinery.

TestAdamicNullishAssertionsAreRefused is not defended. It covers 44 .a exactly-nullish cases across default parameters, returns, arithmetic, templates, assignment, callbacks, arrays, fields and conditionals. The prior present-type subsumer is a .ts control, so the target covers refusal blocks it does not. Its differential diagnostic behavior is nevertheless also asserted by the readiness row and the general 0.1-refusal row. N1 changes the repair text; N6 changes the refusal description; N7 flips the admitted source extension. Each was a production-only attempt at the asserted diagnostic/admission contract, each failed this row, and each had other catchers. These sampled attempts do not prove redundancy for all syntax contexts. Its name does match its assertions; no name/assertion gap was found. No deletion is recommended by this finite result.

TestOptionalWideningCensus is not defended. It shares all covered production blocks with the remaining package (zero exclusive blocks), but uses upstream project options and produces a relation inventory even when other policies would stop lowering. It was enabled using the exact audit census fixture, not left skipped. N3 flips optional-field detection and changes the inventory from one site to zero. N4 empties the reported property name and leaves one site with Property="" instead of "y". N5 reverses source and target arguments and changes one site to zero. It passed all three. Other negative relation tests caught these mutants; the census row did not. These are behavior-changing mutants, not empty-answer probes. Its name suggests a useful census, but assertions do not validate completeness, counts or reported metadata. It serves as a report driver with build/encoding checks rather than an inventory-accuracy gate on this fixture. No diagnostic oracle, test, encoder, or fixture was weakened.

Scope and validation

Clean full baseline passed in 30.332 binary seconds with census files=1 sites=1. Scope contains 37 additions since the audit, all included. Skipped: TestOriginalCycleLedger and the existing Graph array-member subcase pending views-v3. All three requested rows ran in every matrix. No bounded narrowing, production panic, or timeout occurred. Every final standalone N1-N7 diff applies to the starting commit and passed go vet ./internal/lower/ while applied. Each used its own ADAMIC_BUILD_CACHE_DIR. Logs, profiles, scope, census before/after outputs, commands, and full passed lists are included. Restored source control log is included. No other packages were tested and no repository-wide uniqueness is claimed.

Coverage commands use exact target/subsumer names, -coverpkg=./internal/lower, -count=1 and -timeout 90s. For Census the comparison profile uses -run '^Test' -skip '^TestOptionalWideningCensus$' over the rest of this package. Raw profiles and coverage-differences.json record all exclusive blocks.

Brief issues and costs

1. /tmp is only 8.8 GB total, so 15 GB free cannot be attained. Only the earlier unit's /tmp/defend-lower-taste scratch/cache was removed. /workspace had 19 GB free. Neither the repository nor tools were deleted.
2. Census requires unspecified OPTIONAL_WIDENING_CONFIG/OUTPUT paths. The audit's same one-file strict project was supplied. Verdict covers that input, not a larger upstream project.
3. Initial automatic approval review misread 'For each row ... up to three mutants' as three total and rejected the seven-mutant batch. Resubmission explicitly allocated three refusal, one erasure, and three census attempts and was approved. Nothing remained blocked.
4. Initial N2 dropped both entire fast-path blocks and failed go vet because weakOperand became unused. It was restored and corrected to drop only the two returns. That draft has no kill and no replay diff; its failed vet log is preserved separately. Only the compiling correction supports the verdict.
5. N2 is a composite of two return-statement drops because either remaining path still erases 0!. The brief does not specify whether related two-location changes must be counted separately. The exact diff and both origin lines are explicit.
6. Prior audit and current main differ; 37 new top-level tests were included rather than assuming historical scope.
7. The standalone runtime witness initially used plain node, which cannot resolve generated code's adamic import. Both runs were repeated with the repository's unchanged standard oracle/node.mjs runner and printed 0. Generated before/after JS and output logs are included.
8. The final JSON enum omits 'subsumed' and 'twin' mentioned in prose. Neither ambiguity changes these verdicts: rows use defended/not defended and have no executor twin.

Warm tool setup skipped; npm ci in stage3/api completed before baseline; nproc=5. Per-mutant (id, passing rows, failing rows, combined vet/build/run wall seconds): {totals}. Individual binary timings remain in JSON logs; pure build time was not inferred. Baseline, coverage, seven full matrices and restoration remained within the unit budget.
'''
(out/'report.md').write_text(report);print(json.dumps(rows,indent=2))
