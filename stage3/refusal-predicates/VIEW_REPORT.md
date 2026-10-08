dfced0692dc60eeb9bc6db05a4094eaa596c280d: checked predicate views cover 120 of 122 refusal roots.
Commits: views merge 684dc1dc; checked non-null merge 0ac2a1f5; predicate lowering dfced069.
Commands: focused lowering and oracle checks pass; both original reductions emit C without a named stop.
Mutants: admission, .a boundary, writable restriction, and merge pointer rebasing all fail their fixtures.
Limits: two emitNode writable-view refusals remain; the shared counts refresh fails on merged baseline fixtures.

This supersedes the coverage conclusion in REPORT.md. The foundation report and its historical mutant evidence remain unchanged.

The branch is codex/notyet-predicates. The resolved views tip was 167e5fdb779e8c6a64558a3c96b768b7067e776b; the requested checked non-null tip was c41c0e062e99da37820f822968d4df1b48cdaee7. Each merge was committed and pushed before continuing. The feature commit above was also pushed before this report.

The rule reuses the shared view admission in interface_cast.go and lazy contracts in view_lazy.go. It registers the target member family for aliases, then routes concrete narrowed identifier/property reads through the existing structural cast admission. Union sources and structural intersection targets reuse existing descriptors. It adds no new runtime contract. The .ts hatch requires a body, an ordinary predicate parameter, a structural object representation, assignability, and the same writable-slot restriction as interface casts. Body proof remains the first choice. Unproven .a predicates, bodyless declarations, callable values, nominal targets, and incompatible writable views retain their gates.

Ten authored .a witnesses are registered by predicate_views_test.go as mixed-mode tests. Their .a check removes the earlier node! expression so the assertion specifically observes a predicate refusal, then their .ts copy retains node! and exercises the merged checked non-null path. The witnesses cover direct and callback Expression narrowing, Statement callbacks, intersection targets, and union sources. Each valid value agrees with Node in JavaScript, release native, and sanitized native. Each invalid value has a numeric label: Node prints 1 and 42; all Adamic backends print 1, then exit 70 at the label read. The separate writable-view witness runs on Node and stays refused in both .a and .ts. New executable mixed-mode counts are measured by the owning test, recorded in internal/oracle/counts.md, and checked against its rows.

The exact raw CSV tuples from d35a81d36fdafccf827bad0f572d311b2a0d4deb match all 122 ledger rows. All eight body hashes in its TABLE.md match the adapted input. See view-coverage.csv, view-baseline.csv, and view-evidence/input-verification.log.txt. Disabling this hatch on the merged baseline admits one return root, checker.ts:35570:9. The rule adds 119, for 120 after integration:

| Exact refusal family | Admitted | Total |
|---|---:|---:|
| callback isExpression, parameter test | 80 | 80 |
| return on node | 21 | 23 |
| callback isStatement, parameter test | 19 | 19 |

This is a refusal-root audit using the stock checker on the unchanged census input. Its formatter module supplies locations, not the full compiler IR or its read demand. It does not bypass production loading, produce partially executable IR, or claim 120 successfully compiled source sites. Supported witnesses separately demonstrate the runtime rule. Actual compiler bodies can encounter further lowering or member-read stops.

The two remaining rows are transformers/es2015.ts:1349:9 (CapturedThis) and :1354:9 (SyntheticSuper). Their structural source and target are assignable, but widened(target, source) refuses readonly emitNode: EmitNode & { autoGenerate: AutoGenerateInfo } seen through the broad writable EmitNode | undefined slot. This is the existing writable-view design restriction. Admitting them requires a ruling permitting that checked writable relationship or a source adaptation that removes it. The implementation does not weaken the rule.

The original reductions were replayed after both merges and final lowering:

- go run ./cmd/adamic c internal/oracle/testdata/predicate_callback_contract.a: emitted C, no next stop.
- go run ./cmd/adamic c internal/oracle/testdata/predicate_helper_return.a: emitted C, no next stop.
- go run ./cmd/adamic c /tmp/predicate-view-replay.ts, copied from view_valid.a: emitted C, no next stop; its runtime witness passes the three backend modes.

The emitted C byte counts and hashes are in view-evidence/replays.log.txt. These are small reductions, not a claim of complete lowering of the original giant bodies. The census replay command deliberately skips refusal_scan findings, so the exact refusal coverage above uses the raw CSV audit instead.

Final commands and observed outputs, all redirected to files:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go test ./internal/lower -run '^TestPredicateContractRefusals$' -count=1
# PASS, 0.158s
PREDICATE_CENSUS_ROOT=/tmp/notyet-predicates-adapted \
PREDICATE_CENSUS_OUTPUT=/tmp/predicate-view-coverage.csv \
go test ./internal/lower -run '^TestPredicateViewCensus$' -count=1 -v
# PASS, 17.448s; matched=122 admitted=120

go test ./internal/oracle \
-run '^(TestPredicateStructuralViews|TestPredicateWritableViewStaysRefused|TestPredicateCheckedNarrowing|TestPredicateOpenContractsStayRefused)$' \
-count=1 -v
# PASS, 9.398s; ten new mixed-mode witnesses, writable refusal, and prior checks

go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts
# FAIL, 88.085s; full log retained, no successful shared refresh claimed
```

The mandatory counts refresh fails on imported .a non-null fixtures, nbody and maybe-number representation cases, FS cases, and other merged fixtures. Predicate overload tests imported from views also contain .a non-null assertions and conflict with the requested .a refusal. The compiler-area non-null initializer test expects a var refusal while the merged views stack supports that declaration. Those imported assertions were not weakened. The seven non-null-owned count rows were measured and refreshed during the non-null merge; the ten predicate-mode rows are measured and asserted by their owning test. The full shared refresh remains a limitation.

Focused merge checks covered IR/native/JavaScript view and typed-array tests, the flow array-hole exception test, predicate and parameter-property regression fixtures, graph map/set probes, freshness FS/Buffer probes, checked non-null lowering/oracle cases, the explanation driver, and the non-null-owned counts. Their logs are in views-evidence. No full package run or full gate was invoked. An initial broad predicate test selection exposed the imported .a overload conflicts described above; an accidentally broad flow single-assignment fixture sweep exposed the dead-boolean representation merge bug, which was fixed and replayed through its fixture.

| Final mutant | Command selector | What caught it |
|---|---|---|
| Skip predicate family registration and narrowed-read admission | TestPredicateStructuralViews/.*invalid$ | JavaScript exits 0 and prints 42 instead of stopping at label; sanitized native diagnoses the invalid read |
| Allow .a through the hatch | TestPredicateStructuralViews | Each .a predicate refusal assertion observes nil instead |
| Remove writable restriction from predicate eligibility and shared structural admission | TestPredicateWritableViewStaysRefused | .ts compiles instead of remaining refused |
| Skip rebasing map inline entries after graph adoption | TestFreshWriteProbesUseRegions/map_set.a | Node disagreement and native runtime failure |

Each final mutant exits 1, source is restored, and the predicate fixtures pass afterward. The writable mutant initially survived a weaker witness because an earlier assignment refusal masked the target gate. The witness was reduced further, its assertion now requires a predicate refusal, and the rerun fails with '.ts must preserve writable-view refusal: <nil>'. Historical foundation mutants remain reported in REPORT.md; they are not described here as new runs.

Source edits outside predicateRefusal are named in the feature commit: predicates_contracts.go, the new predicates_views.go and census test, the two own oracle test files, counts.md, and eleven .a sources. The two merge messages enumerate their resolution files. A necessary merge reconciliation adds adoptGraphMap in internal/native/graph_regions.go and uses it from emit_expressions.go so inline map entries follow graph relocation. Runtime review: internal/native/runtime/adamic.h combines inherited declarations; no new C runtime helper and no authored C source change were introduced for this predicate rule. No cohere code was copied.

Toolchain setup was performed at the beginning of this unit. The first build failed with undefined censusFieldSlotless, parameterProperty, and classMembersWithParameters in the overlapping compiler checkout; after merge resolution the retry succeeded. Its exact failure is retained in [setup.log.txt](../notyet-predicates/evidence/setup.log.txt), and retry output in [setup-retry.log.txt](../notyet-predicates/evidence/setup-retry.log.txt). nproc=5, CPU quota=4. Retry timing lines:

```text
setup: go ready (0.038s)
setup: node ready (0.051s)
setup: submodules ready (0.146s)
setup: markdown dependencies skipped; step-duration=0.019s
setup: markdown dependencies ready (0.162s)
setup: clang ready (0.392s)
setup: go build ready (69.677s)
setup: test binaries deferred (69.812s)
setup: build cache warm (69.814s)
setup: done on 5 processors (69.904s)
```
