# Lowering chain fixes: remaining subset

Namespace/module initialization now throws catchable TypeError or ReferenceError in both emitted backends; uncaught language errors exit 1, preserving stdout.
Pushed prefixes: aea879601 (items 2/3/4), 91b0148f0 (loader item 6), 1eaed07b7 (lane baseline repair); remaining changes are in this commit.
Focused native/JavaScript Node comparisons, ASan/UBSan, terminating-fixture leak checks, WASI, build and vet pass; final counts regeneration passes in 80.222 seconds.
Five final production overlays and all thirteen original loader mutant classes are caught; the additional loader old-selection mutant is caught too.
The existing typeaware omnibus and its volume C product exceed the 90-second limit; their loader error is gone, but full volume acceptance is incomplete. No full package gate was run.

## Items

| Item | Change and evidence |
| --- | --- |
| 1 | namespaceReadyReads throws TypeError through ir.Throw/adamic_error_new_kind. Native lexical reads/writes construct ReferenceError; runtime gives it Error ancestry. Newly inserted readiness checks propagate MayThrow after readiness proof, and flow.CanThrow includes those read/write edges, preserving exception cleanup. A hoisted namespace var whose undefined value cannot fit number retains its exit-70 representation guard. ModuleNamespaceReadsMatchNode and its readiness mutants pass; new namespace and TDZ catch witnesses compare source Node, JavaScript, native release, ASan/UBSan and leak checks. Explicit WASI selections cover all five unknown-before namespace fixtures, both cyclic module reads, and the scanner premature-value probe. |
| 2 | Already pushed: sorted/nodearray/template are ordinary agreement fixtures. Final explicit native and WASI replays pass. Their witness entry bodies are exercised exactly as written; no claim that an uncalled helper executes. |
| 3 | Already pushed: TestWASIRequestThrows pins exit 1 and empty stderr, as docs/step-21-exceptions.md rules at lines 374-378 and 463. Raw Node supplies an exit-1/name-message stderr control. The ruled Adamic contract deliberately excludes Node's uncaught renderer; adamic_uncaught is unchanged and flushes through adamic_output_flush. |
| 4 | Already pushed: optional-chain RegExp fields/group storage keep the saved receiver and checked field reads. Original regression and new Node witness pass again. |
| 5 | Selector and YAML multiple-value push gaps are closed by the already-landed spread-shorthand member. GAPS records and stale gap expectations are updated; selector's serial push workaround is removed, retaining prefix/index evaluation order. Port controls compare Node, JavaScript and sanitized native; the restored push refusal is caught. |
| 6 | Already pushed: project-aware loading applies only to a project with references. Ordinary mixed .a/.ts roots ignore ordinary tsconfig ownership, as main does. Focused mixed-file type/error controls and Node agreement pass; 13 loader mutant classes plus the restored old-selection mutant are caught. Typeaware drivers were not changed. FactsDecoderGuards passes on chain (14.60 s) and main ea41d314 (3.19 s). TestProduct_profile_controls_lowered, the available equivalent controls product, passes in 45.16 s. TestProduct_corpus_lowered is absent on this base. The full omnibus and later volume C product time out at 90 s after loading succeeds; the later stack is in native constructionNeeded's whole-program walk during emission. These are not reported as passing. |

## Other gate expectations

- checked_any spec fixtures are individual .a files, not a missing main.ts project: pushed loader enumeration repair passes.
- assignment-proofs now has a stage3 top-level owner (5.02 s final replay). Manifestless generics, iteration and iteration-dispatch retain their internal/oracle owners; the directory census names those owners. Explicit native oracle selections hold admitted generics and iteration-dispatch programs to source Node in both backends with sanitizers. Step20IterationOutcomes now does the same, with release and leak checks (18.39 s).
- Step 21 already admits structural Error throws; the stale stop is replaced by a Node agreement test. Existing generator/delegated-generator outcomes are admitted by the landed generators member; only their outcome/reason records change, and both gain ordinary fixture registration.
- objects/25_map_generator.a moves Refused -> NotYet: generators are admitted, but a built-in iterable object still needs a protocol adapter. predicates/10_nullish.a remains Refused: checked-any's earlier explicit-any refusal supersedes the old predicate proof refusal. Both stage3 Node records stay byte-for-byte in the diff. No Compiles fixture regresses. Exact Node/stage0 leaf replays pass.
- Fifteen admitted language TDZ source pins and four corresponding dependency pins leave oracle/node.mjs's terminal convention. Placeholder/representation checks stay terminal. The still-NotYet cycle 06 import-order fixture retains its recorded Node evidence and terminal pin. Source/dependency hash admission tests continue to reject changed source graphs.

## Counts and mutants

Final-existing-counts.md lists all 19 changed existing rows: six lexical reads, one iterator rest TDZ, five namespace TypeErrors, three captured TDZ reads/writes, four switch TDZ reads/writes. Five rows are added. No unexplained row remains. Counts were generated on Linux; an intermediate regeneration exposed the hoisted-var guard and was superseded after restoring that guard.

Final restoration overlays: namespace-old-panic 16.17 s; lexical-old-panic 16.36 s; multiple-push-old-stop 13.39 s; structural-throw-old-stop 14.86 s; readiness-old-effects 15.82 s. Each fails its intended runtime/admission assertion, never just compilation. Overlay sources end in .go.txt, and runners have 120-second process bounds and 90-second Go bounds. The first subset's element-access, uncaught-panic and old-RegExp-chain overlays are also caught. See loader-mutants.json and item-6-report.md for every loader class and the invalid preliminary experiments excluded from the tally.

The added namespace/TDZ/structural Error test leaves finish in 0.52/0.51/0.63 seconds in the final readiness replay; first-subset element-access tests finish in 1.25/1.10/1.25 seconds, regex in 0.80 seconds, mixed-file load test 0.25 seconds and mixed-file oracle package 0.519 seconds. GOMAXPROCS=4; nproc=5, CPU quota=4. No new or changed top-level test takes 60 seconds. Counts is an unchanged existing leaf and is explicitly bounded and monitored.

Commands and complete fixture leaf durations are retained in JSON test output under this review directory. Every command has an outer timeout or monitored session; Go tests have -timeout 90s, except the existing counts regeneration's 210-second bound. WASI uses WASI_SYSROOT and ADAMIC_ORACLE_WASI without replacing native clang in PATH. Build and vet ./internal/... plus the affected stage1/stage3/cmd packages pass. No wrong-output-with-exit-0 or sanitizer failure remains in the focused acceptance set.

Runtime constraints: adamic_uncaught still flushes through adamic_output_flush and does not render the payload. The new reference_error_name is immortal static storage and never mutated; docs/runtime-statics.md does not exist on this base. Runtime's separate stack makes adamic_thrown and adamic_exception_pending _Thread_local; V3 array_holes.c's RangeError assignment must become adamic_error_new_kind when those stacks meet. The existing classes, constructors, absent-message shape and private empty_message slot are preserved.

Setup timings and the initial /tmp space failure/workspace retry are recorded in subset-1-report.md. Targeted cleanup removes only this unit's mutant/main-comparison compiler-cache products, not production cache entries. Initial wrong-compiler-path runs, empty regex selections and bounded timeouts remain evidence of failed attempts and are not counted as acceptance.

Required lane checks pass: `lane checks 2.7 s: gofmt and tools on 241 Go files, t.Parallel on 18 test packages; vet 18 packages`. The separately bounded go vet ./internal/... plus affected stage1/stage3/cmd packages also passes.
