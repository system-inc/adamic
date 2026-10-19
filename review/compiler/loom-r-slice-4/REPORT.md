Re-cut generic body relations, the generics scout, generators, iteration and project references onto the requested fixed main.
Source 2a15510921879ede7590c2272156ca3452cd39ed; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is supplied in the handoff.
Focused tests, Node comparisons with sanitizers, flow corpus, reader guard, structural guards, declared-consumer checks and vet are recorded below and in the JSON evidence.
All 49 recorded mutants were caught in this re-cut; mutants-summary.json lists every name, catcher and current log separately from historical source evidence.
Not covered: whole-package suites, the full gate, ahra/Loom submission, opt-in generic-body census, WASI, admission census and benchmarks.

This branch applies only the source commit's own parent-to-commit patch using git cherry-pick --no-commit. It does not compare main to the source tree. The explicit fixed base overrides the general start-from-current-main instruction: later origin/main movement is intentionally ignored. No cohere code was copied. Setup restored the base's pinned submodule.

Five conflicted hunks in four files were resolved as follows. conflicted-hunks.json preserves each original side; resolved-conflicts.patch records the resulting changes.

1. cloud/fast-gate/compiler-dependencies.json: keep main's stage1/cohere/executable and stage3/lane declarations and add source's stage3/project-loader/probe dependency on internal/load.
2. internal/ir/call_targets_guard_test.go: retain main's explicit leakcheck GoExports input and explanatory comment, adding source's oracle input. Retain source's oracle reader ownership declarations.
3. internal/lower/exceptions.go, throwsOut case list: retain main's ir.ParallelMap and add ir.IteratorMethod and ir.IteratorField. Both parallel callbacks and protocol accessor lookups can throw.
4. internal/lower/exceptions.go, libraryFailure case list: the same union preserves both failure paths.
5. internal/lower/refusals.go: remove both blanket yield and await entries. Source implements synchronous generators; main implements async/await. Keep main's async-generator refusal and its async/parallel preflight, together with source's generic-body and intrinsic-iterator checks. This is the conservative reading that preserves both supported lowerings and their remaining refusals. Focused async/parallel controls pass.

Counts were restored exactly from the fixed base before the single Linux clang 20.1.8 -update-counts invocation. No hand merge was performed. counts-attribution.json contains every changed row and its six before/after counters, with equality to the source counts checked. Counts are regenerated only after compilation and own fixture checks pass.

Neither the source's own patch nor this re-cut changes any internal/native/runtime C or header line. runtime-comparison.json records both empty diffs. Thus there are no changed runtime lines to clear. The source's Go native emitter changes are retained.

Test policy: GOMAXPROCS=4 on a four-CPU quota (nproc reports 5). Every command has an outer timeout or Go test timeout, and test output is written directly to a file. test-times.json lists every observed test and subtest's seconds and result. The focused test selector, corpus runner, fixture runner and command/result JSON files give exact commands. No test was deleted or weakened, no over-budget own test was rerun, and no test split was needed. The complete flow corpus passed, including its remainder for new top-level oracle fixtures; no fixture relocation was necessary.

Structural checks cover Program-held SortedMaps, exact runtime-set witnesses, literal native-agreement registration, parallel-test declarations and call-target reader ownership. Products retain main's placeless infrastructure. The new compiler consumer is declared. Main async and parallel controls, native closure ownership controls and JavaScript mixed-union controls are selected in consumer-controls.jsonl.

The stock TypeScript checks run verify.py and verify_types.py with the pinned stage3/api TypeScript and Node. project-sources/RESULT.json and project-types/RESULT.json retain their observations. Generated adapted TypeScript and binaries are local scratch and are excluded from this commit.

Vet: timeout 150 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage3/project-loader/probe exited 0, with no output. TestCallTargetReaders used -timeout 90s and passed in 15.24 seconds, including export setup.

Setup: GOPROXY=https://proxy.golang.org|direct; ADAMIC_GOCACHE_OFF=1 timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh; unset GOCACHEPROG. Timings in seconds: node .108, Go .158, markdown .442, clang .692, submodules 2.466, cache off 2.468, build 47.869, cache warm 48.016, done 48.039. Go 1.27.1, clang 20.1.8, Node 24.19.0. No setup failure or workaround was needed.

Mutant harness adaptation: main's nodeTypesIndex now returns a rooted path. The standalone-node-prelude mutant changes append(checkRoots, index) to append(roots, index), preserving the original mutant's wrong-root-list behavior. No absent mutation site or compilation failure is credited as a kill. Mutants use non-compilable .go.txt snapshots with go overlays, leaving the delivery tree unchanged. Source historical reports and evidence remain part of its own patch; new raw logs stay local, while current JSON receipts and mutant sources are committed.

Observed counts result: 136.943 seconds aggregate, with all test leaves under 60 seconds. Exactly 55 rows are added and four existing rows change; all 59 after-values equal source slice 4. Existing rows (allocations/frees/retains/releases/peak/regions):

- internal/fresh/testdata/regexp_tree.ts: 91/91/81/85/35/0 to 91/91/85/89/35/0.
- internal/oracle/testdata/regexp.a: 437/437/382/417/58/0 to 437/437/383/418/58/0.
- internal/oracle/testdata/user_iterators.a: 663/663/455/904/64/0 to 1050/1050/677/994/97/0.
- internal/oracle/testdata/user_iterators_rest_tdz.a: 5/0/5/3/5/0 to 11/3/7/3/8/0.

The regexp rows change through boxed-field ownership in generator results; user_iterators rows change through cached protocol dispatch and close ownership. Added rows correspond to the slice's generators, iteration, generic and project-reference fixtures. No unrelated row moves.

Native Go emitter files changed against the fixed base: internal/native/emit_expressions.go, internal/native/iteration.go, internal/native/reuse.go. These are not runtime C files.

Final ordinary test receipts: 3,428 result events, maximum 41.93 seconds. Reader export product: 0.71 seconds. Exactly two opt-in tests skipped: TestGenericBodyRelationsCensus (no GENERIC_BODY_CONFIG) and TestOriginalCycleLedger (no census input). Packages with no test files were compiled by the declared-consumer command. No claim is made to have executed their non-existent tests or either opt-in census. All other selected ordinary tests passed.

The source includes its verbatim Test262 license, whose trailing spaces are retained. Patch whitespace checks pass excluding that unchanged imported license text.

Recorded mutants and catchers:

| Member | Mutant | Catcher |
| --- | --- | --- |
| generic-body-relations | drop-body | dependent body semantic assertion |
| generic-body-relations | drop-witness | resolved witness semantic assertion |
| generic-body-relations | drop-mapper | indexed mapper acceptance assertion |
| generics-scout-main | numeric identity becomes 17 | Node stdout in native and JavaScript |
| generics-scout-main | optional number becomes undefined | Node stdout in native and JavaScript |
| generics-scout-main | array callback optional number becomes undefined | Node stdout in native and JavaScript |
| generics-scout-main | unique root double count | audit root assertion |
| generics-scout-main | binder selection admits concrete names | audit scoped binder assertion |
| generics-scout-main | unescaped union markdown | audit refusal table parser assertion |
| generics-scout-main | missing enclosing binder parents | independent parameters assertion |
| generics-scout-main | changed overlap allowed | resume overlap assertion |
| generics-scout-main | moved site treated as retirement | retirement root assertion |
| iteration-main | next reread | Node stdout in both backends |
| iteration-main | completion value enters body | Node stdout in both backends |
| iteration-main | close skipped on break | Node stdout in both backends |
| iteration-main | close on next throw | Node stdout in both backends |
| iteration-main | close throw replaces original | Node stdout in both backends |
| iteration-main | close skipped on outer continue | Node stdout in both backends |
| iteration-main | snapshot-view | Node stdout |
| iteration-main | strong-weak-slot | ASan SEGV |
| iteration-main | allow-intrinsic-write | intrinsic iterator refusal assertion |
| iteration-main | widen-mutable-element | nested mutable slot refusal assertion |
| iteration-main | wrong snapshot reason | exact NotYet snapshot assertion |
| iteration-main | element-access-old-stop | indexed NodeArray acceptance assertion |
| project-references-main | drop-transitive | TestProjectReferencesTransitive: transitive |
| project-references-main | declarations-only | TestProjectLoaderReferenceSources: TS6305 |
| project-references-main | allow-conflicts | TestProjectReferencesConflictingOptions: want |
| project-references-main | drop-reference-audit | TestProjectReferencesAudit: admitted |
| project-references-main | duplicate-roots | TestProjectReferencesDiamond: duplicate |
| project-references-main | physical-prelude | TestProjectReferencesSharedPrelude: physical prelude console |
| project-references-main | allow-cycle | TestProjectReferencesCycle: circular |
| project-references-main | entry-types-only | TestProjectReferencesTypesUnion: ambient types union lost |
| project-references-main | audit-entry-types-only | TestProjectReferencesTypesUnion: audit lost ambient types union |
| project-references-main | skip-union-declarations | TestProjectReferencesTypesUnionDuplicateSkipped: want both checker duplicate declarations |
| project-references-main | audit-skip-union-declarations | TestProjectReferencesTypesUnionDuplicateSkipped: want ordinary duplicate declarations |
| project-references-main | project-node-roots | TestProjectLoaderNodeImportsKeepReferenceRoots: discarded referenced |
| project-references-main | standalone-node-prelude | TestProjectLoaderStandaloneNodeKeepsPrelude: lost prelude |
| project-references-main | ordinary-project-old-selection | mixed file build assertion |
| generators-main | skip-finally | stdout differs |
| generators-main | delegate-return | stdout differs |
| generators-main | defer-default | stdout differs |
| generators-main | first-next-argument | stdout differs |
| generators-main | forget-pattern-parameter | want the retained whole parameter's cycle refusal |
| generators-main | admit-cycle | LeakSanitizer |
| generators-main | borrow local across suspension | ASan heap-use-after-free |
| generators-main | borrow parameter across suspension | ASan heap-use-after-free |
| freshness reconciliation | field-pure-load | TestIteratorFieldAccessorEscapesReceiver: receiver escape assertion |
| freshness reconciliation | method-pure-load | TestIteratorMethodAccessorEscapesReceiver: receiver escape assertion |
| reader audit reconciliation | unowned oracle operand | TestCallTargetReaders: unapproved call-target read |

All seven external mutant groups passed their catcher assertions: generic 3, generator 6, loader 14, iteration 4, extra 3, reconciliation 3, scout 6 (plus four clean controls). Ten embedded mutants passed their expected-catcher tests: numeric/optional 2, protocol dispatch 6, borrowed suspension frame 2. Total: 49. No compiler-build error or timeout was credited. External mutant test timings are in mutant-test-times.json; the maximum is 8.72 seconds.
