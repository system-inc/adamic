Built: slice 2 re-cut on main; all six original over-budget leaves split, with every assertion retained.
Commit SHAs: main 118885dd60fb28db126bddfc55a2d861beea5e26; source 0a757e0ce28a55fa82797cfc46f68211972e4cdd; one delivery commit, whose SHA is in the final response and git log -1.
Commands and outputs: 74 oracle leaves, 74 support leaves, four readers, three self-compare fixtures, and all eight preparation cases pass; vet passes; counts add 46 owned rows.
Mutants: all 17 originals verified across the prior and finish runs; the 17th now fails the Node comparison with a narrowed-union stop. New ownership and reader guard mutants also fail.
Not covered: full package suites, full gate, WASI, and macOS. The original first-green-within-30-minutes target was missed during cold setup.

The branch starts at the fetched main tip above. Main moved once during this finish; the delivery commit was rebased once. A final fetch confirmed that main had not moved again. The original squash had no textual conflicts. The rebase conflicted only in internal/oracle/counts.md, which was restored from main and regenerated once on this base. Other overlapping edits merged automatically, preserving both sides.

The two explicit Weak lifetime counters use main's keyed countCase registry. Counts move from 1037 to 1083 rows: 46 additions, zero changed existing rows, zero removals. Owners: callback-widening: 35, miscompile-fxspptb-2a: 3, miscompile-fxspptb-2b: 5, self-compare-main: 3. Every row and its values are recorded in counts-attribution.json.

The runtime diff is exactly slice 2's own Error message repair in adamic.h (+2), exceptions.c (+15/-2), and object.c (+2). Its zero-context patch matches the pre-rebase slice after normalizing hunk locations and blob IDs. Main's runtime edits are retained. runtime-own-diff.patch records the complete diff. No code was copied from cohere.

## Splits and every new leaf's seconds

Single-fixture oracle leaves are split by backend; sanitized cases retain leak checks. Self-compare is split into closure, array-fill and sweep fixtures. The fixture ownership guard still requires every manifest row exactly once, and rejects missing, duplicate, mixed or nonexistent owners. Reader checks scan production, internal tests and external tests separately for each of native, lower, fresh and flow, preserving metadata and stale-entry assertions.

Export preparation is split again into key discovery and archive cases. Serial key discovery avoids four simultaneous listings; parallel archive cases retain only Go's __.PKGDEF member, which go/importer reads, using the original ar envelope. Test-variant import aliases are preserved. Main's normal precomputed input index remains supported, and every archive case prepares itself when selected alone. All final preparation cases include setup in the seconds below.

| Test | Seconds |
|---|---:|
| TestCallbackWideningNumberEveryRelease | 25.37 |
| TestCallbackWideningNullableObjectMapJavaScript | 25.59 |
| TestCallbackWideningNullableObjectMapRelease | 26.06 |
| TestCallbackWideningNullableObjectMapSanitized | 27.41 |
| TestWeakDifferentialHeldUnionJavaScript | 0.92 |
| TestWeakDifferentialHeldUnionRelease | 1.31 |
| TestWeakDifferentialBooleanFilterJavaScript | 2.08 |
| TestWeakDifferentialHeldUnionSanitized | 4.02 |
| TestWeakDifferentialBooleanFilterRelease | 1.63 |
| TestWeakDifferentialBooleanFilterSanitized | 3.73 |
| TestCallbackWideningNumberEveryJavaScript | 0.43 |
| TestCallbackWideningNumberEverySanitized | 2.73 |
| TestFixturesSelfCompareArrayFill | 40.68 |
| TestFixturesSelfCompareClosure | 40.79 |
| TestFixturesSelfCompareSweep | 40.94 |
| TestCallTargetExportInputsNative | 10.38 |
| TestCallTargetExportInputsLower | 10.45 |
| TestCallTargetExportInputsFresh | 19.16 |
| TestCallTargetExportInputsFlow | 16.88 |
| TestProduct_CallTargetReaderArchivesNative | 25.82 |
| TestProduct_CallTargetReaderArchivesFlow | 26.25 |
| TestProduct_CallTargetReaderArchivesLower | 26.97 |
| TestProduct_CallTargetReaderArchivesFresh | 26.32 |
| TestCallTargetReadersFlow | 27.49 |
| TestCallTargetReadersFresh | 29.83 |
| TestCallTargetReadersLower | 29.91 |
| TestCallTargetReadersNative | 30.04 |

Measured with GOMAXPROCS=4 and a four-CPU cgroup quota. Scripted batches additionally restrict CPU affinity to four CPUs. nproc reports 5. new-leaf-seconds.json contains all 27 replacement and preparation cases; finish-test-seconds.json contains every completed affected leaf. The edited fixture ownership guard also passes.

## Exact finish commands and results

| Phase | Exit | Wall seconds | Command |
|---|---:|---:|---|
| reader-products | 1 | 218.747 | `timeout 240 go test ./internal/ir -run ^TestProduct_CallTargetExports -json -count=1 -timeout=90s` |
| reader-products | 0 | 123.834 | `timeout 240 go test ./internal/ir -run ^TestProduct_CallTargetReaderExports -json -count=1 -timeout=90s` |
| readers | 1 | 70.76 | `timeout 180 go test ./internal/ir -run ^TestCallTargetReaders -json -count=1 -timeout=90s` |
| oracle | 0 | 136.301 | `timeout 300 go test ./internal/oracle -run ^Test(CallbackWidening\|WeakDifferential\|Miscompile2A\|SelfCompare) -json -count=1 -timeout=90s` |
| fixtures | 0 | 103.104 | `timeout 240 go test ./stage3/fixtures -run ^Test(FixturesSelfCompare\|FixtureDirectoriesHaveTopLevelTests) -json -count=1 -timeout=90s` |
| support | 1 | 59.393 | `timeout 300 go test ./internal/native ./internal/ir ./internal/flow ./internal/fresh -run ^Test(CallbackWidening\|CallbackAdapterTargets\|ClosureTargetsBoundOnlyProvenValues\|ErrorUndefinedMessageHasNoOwnProperty\|SplitSelfCompareAgreesWithNode)$\|^TestCallbackWidening -json -count=1 -timeout=90s` |
| reader-products | 1 | 149.286 | `timeout 240 go test ./internal/ir -run ^Test(CallTargetExportInputs\|Product_CallTargetReaderArchives) -json -count=1 -timeout=90s` |
| vet | 0 | 30.953 | `timeout 240 go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/fresh ./internal/oracle ./stage3/fixtures` |
| readers | 0 | 56.614 | `timeout 180 go test ./internal/ir -run ^TestCallTargetReaders -json -count=1 -timeout=90s` |
| counts | 0 | 433.571 | `timeout 660 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts` |

The two IR support checks that did not build in the combined support command were completed with `timeout 180 go test ./internal/ir -run ^Test(CallbackAdapterTargets|ClosureTargetsBoundOnlyProvenValues)$ -json -count=1 -timeout=90s`. Fresh archive preparation was interrupted after about 33 seconds by the combined package's 90-second limit, after the other seven cases passed. It was completed alone with `timeout 180 go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFresh$ -json -count=1 -timeout=90s`, passing in 26.32 seconds. Final TestCallTargetReaders and changed-package vet pass.

All test output was written to files. Tests were focused selectors, not whole package suites or the full gate. Source Node, JavaScript, native release, ASan, UBSan and Linux leak checks retain each fixture's original contract. The Weak freed-target witnesses retain their explicit native/Node lifetime difference.

## Every mutant and its catcher

| Mutant | Observed catcher |
|---|---|
| error-spread | want path-bearing stop |
| maybe-setter | want path-bearing stop |
| long-name | want path-bearing stop |
| computed-name | panic: Unhandled case in Node.Text |
| constructor-capture | want path-bearing stop |
| optional-error | member access within null pointer |
| error-own-message | native , Node  |
| method-views | both backends unexpectedly admit the pinned view NotYet, release: exit=0 stdout="t1 none, release: stdout differs, AddressSanitizer |
| union-liveness | release: exit codes differ, sanitized: exit codes differ, runtime error: member access within null pointer |
| weak-callback | both backends must stop at |
| refusal-revert | TestCallbackWideningRefuseSome, TestCallbackWideningRefuseReduce, TestCallbackWideningRefuseMap, TestCallbackWideningRefuseSort, TestCallbackWideningRefuseFind, TestCallbackWideningRefuseEvery, TestCallbackWideningRefuseForEach, TestCallbackWideningRefuseFilter |
| adaptation-omitted | TestCallbackWideningNumberFilter, TestCallbackWideningNumberMap, TestCallbackWideningNumberForEach, TestCallbackWideningNumberReduce, TestCallbackWideningNumberFind, TestCallbackWideningNumberSort, TestCallbackWideningNumberGenericMap, TestCallbackWideningNumberSome |
| unrepresented-parameter-admitted | TestCallbackWideningUnrepresentedParameter |
| adapter-target-omitted | TestCallbackAdapterTargets |
| refusal-revert | TestCallbackWideningRefuseSome, TestCallbackWideningRefuseReduce, TestCallbackWideningRefuseMap, TestCallbackWideningRefuseSort, TestCallbackWideningRefuseFind, TestCallbackWideningRefuseEvery, TestCallbackWideningRefuseForEach, TestCallbackWideningRefuseFilter |
| unrepresented-parameter-admitted | TestCallbackWideningUnrepresentedParameter |
| adapter-target-omitted | TestCallbackAdapterTargets |
| adaptation-omitted | TestCallbackWideningNumberEveryRelease, TestCallbackWideningNumberEverySanitized |
| static-inline-definition | TestSplitSelfCompareAgreesWithNode: clang undefined-inline diagnostic; structural only, not counted as a semantic mutant |
| split-numeric-fold | TestSplitSelfCompareAgreesWithNode: split native stdout differs from Node |
| weak-callback-isolated | both backends unexpectedly admit the pinned callback NotYet, AddressSanitizer: heap-buffer-overflow, sanitized: exit codes differ |
| weak-counter-registration-omitted | TestCountsAreRecordedTable |
| fixture-owner-omitted | has no top-level test |
| reader-allowlist-omitted | unapproved call-target read |
| nullable-token-omitted | Node stdout/stderr/exit comparison catches the narrowed-union checked stop in both release and sanitized artifacts |
| constant self-equality | TestSelfCompareConstantMutant: successful mutated build and execution disagree with Node for NaN |

The 16 earlier original semantic proofs were accepted on base 199174800998904d8d646dfa5892f512cd5cb20f. The constant self-equality check also ran again on this base. The new Every release and sanitized leaves restore the ninth original adaptation witness; its JavaScript leaf passes that particular native-representation mutant. Together with prior-callback-results.json, all nine original numeric fixtures are covered. The current nullable-token mutant completes the original set of 17.

For nullable-token-omitted, the compiler artifact was built with the recorded lower/callback_widening.go overlay. Node observes stdout `1,null\n`, empty stderr, exit 0. Both native artifacts observe the narrowed-union panic, exit 70. Release build took 7.083 seconds and sanitized build 10.030 seconds; both builds succeeded. ASan/UBSan were enabled for the latter. The byte comparison catches the mismatch; no sanitizer fault is claimed. Exact commands, stdout, stderr and timings are in callback-mutants/nullable-token-omitted/artifact-results.json. The first direct Go-test attempt exhausted its outer limit during cold preparation and has no completed verdict; it is not credited. The artifact script initially expected a UBSan fault, but the recorded bytes show a checked stop, so its assertion and report were corrected without repeating either program.

The ownership guard parses physical source. Its first Go-overlay mutant therefore had no effect on its observation and is not credited. The real source name was then changed temporarily, the guard failed for the missing fixture owner, and the original file was restored in finally. The reader allowlist omission fails for an unapproved read. Both results are in grain-mutants.json; mutant sources are .go.txt.

The earlier union-liveness child completed after its collection supervisor ended; the complete log records FAIL in 9.14 seconds, while its process exit code is unknown. Timeouts and build failures are never credited as semantic kills.

## Preparation and failed attempts

Setup reached its 600-second limit in go list -deps -export -json ./.... Its generated environment was usable and was sourced in later shells. The exact timing lines follow.

```text
setup: node ready (0.265s)
setup: go ready (0.352s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (2.310s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=5.627s
setup: markdown dependencies ready (6.559s)
setup: submodules ready (451.091s)
setup: shared cache ready (492.723s)
```

The prior six budget failures remain historical evidence in overdue-leaves.json. Their replacement checks pass. Additional full-archive preparations timed out; retaining only export members reduced copying, but combined key discovery and archive cases still took 62.62 to 67.23 seconds under parallel preparation. Those cases were split into the eight final preparation cases above; the overdue cases were not repeated. The final combined package alarm interrupted Fresh below its share, so only that unfinished case was completed separately.

I changed reader declarations while the combined support command was compiling; its generated IR test main referenced the preceding names. The other 72 support leaves completed and passed. After declarations settled, the two IR support leaves passed separately, all readers passed, and vet passed. The first per-package reader run also exposed a missing import alias for flow external tests, which was corrected and is covered by the final readers.

Runtime archives were prepared through native.RuntimeRequestFor, RuntimeRequest.Build with four workers, and main's declared buildcache inputs. Go export and fixture-hook preparation ran outside budgeted fixture cases. Preparation commands had hard limits and retained logs. logs.tar.gz bundles raw logs and JSONL evidence. The initial historical reports and scripts remain available under review/compiler/chain-slice-2; checks.jsonl and finish-checks.jsonl distinguish prior and finish attempts.
