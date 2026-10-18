Built: search-shrink #b75jjs3 on current main, preserving callback widening and nullable-reference boxing.
Commit: one commit on compiler/loom-r-search-shrink; base 67733635b0252eb9d3bb3645c72262c7b18361e1; source df57ed18956e74caa7e126ba59225033ad06f015.
Validation: 1,741 current tests accounted for as pass/expected skip after infrastructure recovery; counts, a-check, guards and vet pass.
Mutants: skip-check and old-stop caught in both backends; old diagnostic matcher caught; lost-null conversion caught by Node stdout.
Not covered: full repository gate, opt-in benchmarks/WASI corpus, and an isolated budget measurement for the unchanged moves test (75.75 s after disk recovery).

Branch: `compiler/loom-r-search-shrink`. Main was fetched at the start; its tip was exactly the base above. This is one commit containing the source commit's own delta, not a main-to-source comparison. The original validation at c0c4e10211 was read; source historical reports and overlays remain under review/compiler/search-shrink-chain/ and are not new validation claims.

## Merge and compatibility resolutions

`git merge-tree --write-tree --merge-base=df57ed1895~1 origin/main df57ed1895` produced tree ab96d5ce04bbc8a92790f6d2961554b482288de6 and one conflicted hunk, `internal/native/emit_arrays.go` at the callback invocation and forEach result handling. Main's newer selective closure convention had removed the old discardClosureResult helper. Keep the search change's one-time conversion to SearchFirst and call callbackCall with that argument; keep main's `forEach && !visit.Returns.IsReference()` emission and reference-result release. Do not restore the obsolete void-result helper. Every other source hunk merged automatically, including main's callback widening and predicate representation checks in lowering.

The source counts merge was discarded in favor of main's counts.md and regenerated through the official writer, never edited by hand.

An additional compatibility check found the source direct conversion bypassed main's nullable-reference adapter: the first present null in `(Item | null)[]` passed to an unknown callback was printed as `hole` natively, while source Node and JavaScript printed `null`. `logs/nullable-node.log`, `nullable-native.log` and `nullable-js-with-runner.log` record this before observation. Preserve that distinction in IR (`SearchNull`), computed from checker types rather than a type-name string. For a present nullable-reference slot boxed into a union use the existing `adamic_null` token; a removed index remains undefined. No runtime source changes. `search_shrink_nullable.a` and its top-level test hold this to Node. The review probe is retained as evidence. A raw direct Node invocation of emitted JS initially lacked the Adamic module hook; the authoritative rerun uses oracle/node.mjs.

Assumptions: retain source policy that omitted and defaulted value parameters admit undefined, because no undefined value is consumed as T before default binding. Retain existing NotYet/refusal boundaries for unsupported callback representations. Treat preserving main's nullable callback protocol as necessary compatibility work within this cut. No other array method policy changes.

## Focused behavior and test grain

The 14 original .a fixtures are byte-for-byte source copies (`fixture-bodies.json`). All source programs finish on Node. Four deliberately non-admitting callbacks stop at exit 70 in all compiled modes, naming method, index and element type; other fixtures match Node exactly. Sanitized native uses ASan/UBSan and successful runs are checked for leaks. Both release native and backend JavaScript are compared too.

The source harness called a removed leakSanitizer helper. Adapt it to main's shared `leaksUncached` helper. Four initial focused leaves took 70.89-71.50 s waiting on the general observation cache's unrelated full-runtime identity builds. Rather than rerun those unchanged, focused tests now invoke the same repository Node runner directly, retain all comparisons and use main's leak checker. Their measured final durations, setup included, are:

| Test | Seconds |
|---|---:|
| TestSearchShrinkFindLastIndexMaybe | 6.40 |
| TestSearchShrinkUnknown | 6.41 |
| TestSearchShrinkNumberElementMaybe | 6.46 |
| TestSearchShrinkFound | 6.49 |
| TestSearchShrinkFindIndexChecked | 1.31 |
| TestSearchShrinkReferenceElementMaybe | 1.32 |
| TestSearchShrinkFindLastIndexChecked | 1.41 |
| TestSearchShrinkFindLastChecked | 1.53 |
| TestSearchShrinkDefault | 1.13 |
| TestSearchShrinkOmitted | 1.25 |
| TestSearchShrinkBoolean | 1.37 |
| TestSearchShrinkFindChecked | 1.41 |
| TestSearchShrinkMutantSkipCheck | 1.17 |
| TestSearchShrinkMutantOldStop | 0.93 |
| TestSearchShrinkFindMaybe | 1.24 |
| TestSearchShrinkFindLastMaybe | 1.34 |
| TestSearchShrinkNullableReference | 0.49 |

## Mutants

- `TestSearchShrinkMutantSkipCheck`: both generated programs compile. Native exits 0 with `0 1\n1 2\n2 0\n-1\n`; JavaScript exits 0 with `0 1\n1 2\n2 hole\n2\n`, demonstrating undefined entering a number callback. Both are caught by the required exit 70, not a warning or sanitizer failure.
- `TestSearchShrinkMutantOldStop`: both backends exit 70 after `1\n2\n` instead of Node's successful hole visit. The admitting fixture catches both by exit code.
- `fuzz-old-matcher-overlay.json`: remove only the new typed-search diagnostic alternative. `TestJudgeReadsThePanicLine` fails four classification assertions, expected exit 1.
- `native-without-null-conversion-overlay.json`: disable only SearchNull's conversion. `TestSearchShrinkNullableReference` compiles, both native variants exit 0 without stderr, and both fail stdout against Node (`0 hole` instead of `0 null`). JavaScript still agrees. Expected test exit 1. Overlay source files end in .go.txt.

## Counts

Linux, house clang 20.1.8. First full official update passed in 139.607 s. The subsequently discovered nullable witness was added through a focused invocation of the same official writer. 15 added rows and four modified rows, 19 total. The existing callback-widening searches no longer allocate the extra representation adapter's captured cell and closure because conversion is done at the search call. No unrelated row changed. Columns below are allocations/frees/retains/releases/peak/in-regions.

| Fixture | Before | After |
|---|---|---|
| internal/oracle/testdata/callback_widening/number-find.a | 8/8/7/13/5/0 | 6/6/2/8/3/0 |
| internal/oracle/testdata/callback_widening/number-findIndex.a | 10/10/7/15/6/0 | 8/8/2/10/4/0 |
| internal/oracle/testdata/callback_widening/number-findLast.a | 10/10/7/15/6/0 | 8/8/2/10/4/0 |
| internal/oracle/testdata/callback_widening/object-find.a | 9/9/16/21/7/0 | 7/7/9/14/5/0 |
| internal/oracle/testdata/047cb0d_f_maybe.a | new | 11/11/8/20/5/0 |
| internal/oracle/testdata/047cb0d_f_maybe_ref.a | new | 13/13/16/31/7/0 |
| internal/oracle/testdata/57f2d04_find_last_shrinks.a | new | 8/5/8/13/8/0 |
| internal/oracle/testdata/57f2d04_find_last_shrinks2.a | new | 8/5/8/13/8/0 |
| internal/oracle/testdata/search_shrink_findIndex_checked.a | new | 10/8/5/13/6/0 |
| internal/oracle/testdata/search_shrink_findLastIndex_maybe.a | new | 10/10/5/16/6/0 |
| internal/oracle/testdata/search_shrink_find_maybe.a | new | 13/13/10/24/6/0 |
| internal/oracle/testdata/search_shrink_findLast_maybe.a | new | 9/9/7/17/6/0 |
| internal/oracle/testdata/search_shrink_find_checked.a | new | 10/8/5/13/6/0 |
| internal/oracle/testdata/search_shrink_unknown.a | new | 14/14/9/25/6/0 |
| internal/oracle/testdata/search_shrink_boolean.a | new | 9/9/7/19/4/0 |
| internal/oracle/testdata/search_shrink_default.a | new | 10/10/4/15/4/0 |
| internal/oracle/testdata/search_shrink_omitted.a | new | 4/4/4/9/3/0 |
| internal/oracle/testdata/search_shrink_found.a | new | 15/15/12/32/8/0 |
| internal/oracle/testdata/search_shrink_nullable.a | new | 12/12/12/26/6/0 |

## A-check and runtime clearance

The unchanged README-pinned Gate.aCheck predicate at fbac28c62493f27a788edc02a18bc8edb68de5da runs the final branch compiler on every changed .a (`paths.txt`): **16 checked, 0 NotYet, 0 refused, 0 type errors**. All pass; no headers were added and no source-line expectations move. `final.json` contains every path, diagnostic, output hash and status. No refused-header predicate mutant applies because no touched file requires a header.

**Runtime C/header lines changed: none.** `git diff base -- internal/native/runtime/*.c internal/native/runtime/*.h` is empty, as is the original source commit's runtime delta. The null conversion uses an existing token.

## Full package verification and guards

All **1,741** current top-level tests in the touched package trees are accounted for as passing or expected skips, with no missing or failed test left (`validation.json`, cross-checked against `go test -list`). This is aggregate coverage from full commands and explicit recovery runs, not a claim that the original full commands exited green.

| Package | Passed | Expected skips | Total |
|---|---:|---:|---:|
| internal/ir | 27 | 0 | 27 |
| internal/lower | 329 | 0 | 329 |
| internal/lower/cycleledger | 0 | 1 | 1 |
| internal/javascript | 2 | 0 | 2 |
| internal/fuzz | 38 | 0 | 38 |
| internal/native | 384 | 92 | 476 |
| internal/native/runtimeproducts | 5 | 0 | 5 |
| internal/oracle | 710 | 153 | 863 |

Commands use setup's env.sh, unset GOCACHEPROG, GOMAXPROCS=4, GOFLAGS including -p=1 (initial combined core run used -p=4), and -parallel=4. Every test command has both a hard outer timeout and Go's -timeout. All output is direct to logs.

- Full core command: `timeout 180 go test ./internal/ir/... ./internal/lower/... ./internal/javascript/... ./internal/fuzz/... -count=1 -parallel=4 -timeout 150s -json` -> exit 124 at the overall command limit. The inventory identified 108 unfinished lowering tests; all passed in five batches of at most 24, each with `timeout 120 go test ./internal/lower/... -run '^(names)$' -count=1 -parallel=4 -timeout 90s -json`. The ledger subpackage was run separately and passed. All IR tests passed after the final reader recovery.
- `timeout 180 go test ./internal/javascript/... ./internal/fuzz/... -count=1 -parallel=4 -timeout 150s -json` -> exit 0 (`logs/javascript-fuzz-final.jsonl`).
- `timeout 880 go test ./internal/native/... -count=1 -parallel=4 -timeout 850s -json` -> exit 1 for 22 normalization tests sharing a disk-failed product and one unconditional WASI test whose original process lacked WASI_SYSROOT. The normalization product and all 22 affected tests pass in a new process; the unchanged WASI test and its native/WASI binding mutants pass after installing the SDK (`logs/native-wasi-after-suite.jsonl`). Every one of the 476 native tests and five runtimeproducts tests now has a pass/expected-skip result.
- `timeout 880 go test ./internal/oracle/... -count=1 -parallel=4 -timeout 850s -json` -> exit 1, 809.035 s, only for TestProduct_MovesNative_asan writing its metadata during disk-full. That product passes its explicit recovery run. All 863 oracle tests are now pass/expected skip, including the runtime-set witness guard and all new fixture leaves.
- **Budget observation retained:** unchanged main TestMovesAgreesWithNode_asan passed correctness but took 75.75 s because it rebuilt the product whose preceding metadata write failed. It was not rerun unchanged. This loaded infrastructure-recovery observation does not certify that leaf's 60 s budget. The product is now independently built and green; no movement test was edited to hide the observation. All other current top-level final observations are under 60 s, and every added/touched search leaf is under 60 s. Skips remain main's existing optional benchmark/WASI requirements, not new exemptions.
- Official counts: `timeout 240 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 210s -json -args -update-counts` -> exit 0, 139.607 s. Added nullable row: same writer, `-run '^TestCountsAreRecorded_[0-9]+$/internal/oracle/testdata/search_shrink_nullable[.]a$'`, exit 0.
- Final guards: `go test ./internal/ir -run 'TestCallTargetReaders|TestCallTargetExportInputs|TestProduct_CallTargetReader|TestEveryMapAProgramHoldsIsSorted' -count=1 -parallel=4 -timeout 90s -json` -> exit 0; TestRuntimeSetsAreEachUsed and TestNativeAgreementRowsNameRegisteredPrograms pass in the package inventories; `go test ./cmd/adamic-gate -run '^TestEveryTestIsParallelOrSaysWhy$' -count=1 -timeout 90s -json` -> exit 0.
- `go test ./internal/flow ./internal/fresh -run 'Test(Flow|Fresh)Corpus(Remainder|UnitsCoverEveryProgram)$' -count=1 -parallel=4 -timeout 90s -json` -> exit 0 on the final 15-fixture addition (`logs/corpus-final.jsonl`).
- `timeout 180 go vet ./internal/ir/... ./internal/lower/... ./internal/native/... ./internal/javascript/... ./internal/oracle/... ./internal/fuzz/...` -> exit 0, no output. `gofmt -l` on every changed Go file and `git diff --check` -> no output.
- A-check: `PYTHONDONTWRITEBYTECODE=1 timeout 240 python3 review/compiler/loom-r-search-shrink/check-headers.py final` -> exit 0, 16/16 checked. Behavioral mutants run inside the focused tests; the two overlays are separate expected-exit-1 runs as described above.

All commands use GOMAXPROCS=4 and Go package/test concurrency limits, with hard process and Go test timeouts. Output is written directly to logs. The complete native and oracle suites started before the nullable repair were stopped and restarted on the final compiler (`interrupted-suites.json`). Those interrupted logs are retained, not claimed as final passes.

Guard obligations: no new Program map, compiler-consuming package or runtime feature set was introduced. Run TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy and all four TestCallTargetReaders units. Rerun flow/freshness corpus remainders and their coverage guards for the added oracle files. No reader is removed, and no unrelated package test is changed to make space.

## Toolchain and infrastructure observations

GOPROXY=https://proxy.golang.org|direct; nproc=5, cgroup quota four CPUs. Go 1.27.1, Node v24.19.0, native clang 20.1.8. Initial setup hit its hard 240 s limit while warming Go exports; setup retry with optional duplicate Go cache disabled passed in 103.351 s. All timing lines are in logs/setup-retry.log. Optional WASI SDK setup is recorded separately because an existing unconditional native unit needs it; no test is skipped or edited to avoid that dependency.

The 32 GB workspace filled while copying several gigabyte-sized Go-export products. Preserve all failures in initial logs. Only verified retired products and the inactive optional duplicate cache were removed. Byte-identical immutable export archives were then hard-linked without changing a product's path or bytes. Actions are recorded in cache-reclamation.json and cache-deduplication*.json. A first optional WASI extraction also hit disk-full and was removed as an incomplete installation before retry. Remaining affected test units are rerun explicitly after recovery; no source reader is deleted. A preliminary formatting evidence script omitted setup's PATH; the corrected gofmt command uses env.sh.

Optional SDK retry passed in 182.660 s; SDK link probe ready in 12.377 s, Go build ready as recorded in logs/setup-wasi-retry.log. Native compiler remains clang 20.1.8; SDK 27 uses clang 20.1.8-wasi-sdk. Main's official counts writer retains its current seven-column table schema and puts measurement digests in products.
