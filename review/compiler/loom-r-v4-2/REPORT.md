Built: views V4 carrying V3 plus the two reproduced gate fixes and a budget-safe witness split, on compiler/loom-r-v4-2.
Commit: one commit above exact dbdf977f5b8b0979d414ebe6f20e7aadd312e309; final SHA is in the response.
Commands and outputs: named JSON shards pass on base and V4; both reported guards pass; remaining validation is summarized below.
Mutants: the original source-report mutants and main library-spread preservation mutant are rerun with individual logs and catchers below.
Not covered: ahra loom check/full gate, the five unnamed Golem shard identities, and documented pending/skipped controls.

Started with git fetch origin, created compiler/loom-r-v4-2 from exact dbdf977f5b, and applied b7268d277e0aa78a5def98f1d53093dcc3b471ba with git cherry-pick --no-commit. The newer fetched origin/main is intentionally not the base. There were no new application conflicts. The carried lower.go resolution and runtime/counts attribution remain documented in ../loom-r-v4/REPORT.md.

The naming guard reproduced a global collision between internal/native and internal/javascript TestViewCallableBoxingUnknownProducer. Renamed only those V4 tests to TestNativeViewCallableBoxingUnknownProducer and TestJavaScriptViewCallableBoxingUnknownProducer. The guard and pinned offenders are unchanged. guards-before.jsonl records its refusal; guards-after.jsonl records acceptance, so the original names supply a concrete failing control.

The compiler census reproduced internal/oracle/array_runtime_mutants_test.go:arrayRuntimeMutant directly invoking clang with absolute source and runtime paths. Kept its private mutated runtime snapshot and native.RuntimeLibraryForSource, then passed the generated C, output, options and private archive to native.BuildWithRuntime. Removed the hand-written clang command, flags and source-file write. Main's entry point supplies relative compiler input names, prefix maps and compiler admission. No census exemption was added. The original direct call is the guard's failing control; all four array-runtime mutant tests retain their semantic checks and pass through the new entry point.

JSON observations are in json-comparison.json and base-comparison-<shard>.log / branch-comparison-<shard>.log. All three named shards pass on both trees, including Go oracle, source Node, release native, ASan/UBSan/LeakSanitizer native and JavaScript-backend comparisons. The five other Golem test names were not supplied; conservatively also ran neighbors 2037, 2038, 2039, 2042 and 2043. These are extra coverage, not a claim that they identify Golem's five missing names.

| JSON shard | Exact base seconds | V4 seconds | Result |
|---|---:|---:|---|
| 2036 | 4.66 | 4.16 | both pass |
| 2040 | 3.79 | 4.27 | both pass |
| 2041 | 4.38 | 3.20 | both pass |
| 2037 | 2.63 | 3.18 | both pass |
| 2038 | 3.47 | 4.65 | both pass |
| 2039 | 4.85 | 5.48 | both pass |
| 2042 | 5.91 | 3.84 | both pass |
| 2043 | 4.54 | 3.43 | both pass |

No branch-only failing input was observed, so no lowering reduction or speculative lowering change was made. This does not explain the Golem failure. It establishes local agreement on the named inputs with matching base, compiler and corpus. The final binary tests were run from their package directories in isolated base and V4 snapshots, each with a real pinned cohere worktree; cohere implementation was not copied.

Initial local setup attempts are retained, not counted as input regressions. The oracle owner had to be settled with TestProduct_JSONGoOracle. npm ci --prefix stage3/api provisioned corpus.pin inputs absent on this box. TestProduct_JSONPinnedCorpus then settled each tree's corpus. The original checkout's historical pinned JSON contains /workspace/adamic and the product relocation check rejected it here; an isolated V4 worktree under /tmp avoided that coincidental source-data path. A symlinked cohere checkout initially defeated Go's overlay and produced the stock formatter command without --cases; real cohere worktrees corrected that setup. Both final comparisons use the same b074075f9339 Go oracle product, the same corpus pin 2b4dbde172c1c2f23c065ecae8b906f8fcb72a20, and distinct base/V4 lowering and native products. None of these setup workarounds edits a repository test or weakens a guard.

Commands used for each final JSON shard, output redirected to its individual log:

    source /workspace/adamic-tools/env.sh
    timeout 180 go test -c -o <test-binary> ./stage1/cohere/json
    timeout 120 go test ./stage1/cohere/json/oracles -run '^TestProduct_JSONGoOracle$' -count=1 -timeout 90s -v
    timeout 95 <test-binary> -test.run '^TestProduct_JSONPinnedCorpus$' -test.timeout=90s -test.v
    timeout 95 <test-binary> -test.run '^TestPortMatchesGoCohere_<shard>$' -test.timeout=90s -test.parallel=1 -test.v

No fixture, production lowering or production runtime line changes versus b7268d27. Counts were not regenerated because no fixture or compiler/runtime behavior changes; counts.md is byte-for-byte the prior Linux clang 20.1.8 regeneration. unchanged-from-v4.json records those comparisons. Thus the prior source-own runtime +/- delta remains identical, with zero additional runtime lines requiring clearance.

Setup passed with GOPROXY='https://proxy.golang.org|direct', timeout 300 bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. setup.log retains every timing line: submodules 0.105 s, Go 0.113 s, Node 0.124 s, markdown 0.235 s, clang 0.440 s, shared cache 0.931 s, Go build 15.538 s, tests deferred 15.608 s, build cache warm 15.609 s, done 15.634 s. nproc=5, cgroup cpu.max=400000 100000 (four CPUs); clang 20.1.8, Go 1.27.1, Node 24.19.0.

Validation details and final inventories follow.

Focused validation uses run-focused.py, the original V4 inventory with the two renamed selectors, current main controls and runtime products. It compiles package test binaries with an outer 180 s bound and runs each exact top-level test with -test.timeout=90s, -test.parallel=1 and an outer 95 s bound, at most two concurrently. focused-results.jsonl and selected-focused.json are the inventory; each leaf has its own log.

The earlier direct-reader inventory was rerun with run-readers.py: all 53 changed/direct-reader packages compile; 48 focused reader checks exit zero. Two are explicitly skipped: cycleledger TestOriginalCycleLedger (original source pin unavailable) and estree TestCookedSurrogateLibraryGap (existing closed-gap skip). These are not whole-package test runs. reader-results.jsonl has every build and test time.

Final guards command (final-guards.jsonl):

    timeout 180 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate ./internal/compilercensus -run '^(TestEveryMapAProgramHoldsIsSorted|TestCallTargetReaders|TestNativeAgreementRowsNameRegisteredPrograms|TestEveryTestIsParallelOrSaysWhy|TestEveryTestHasOneDescriptiveName|TestEveryCCompilerRunsThroughNative|TestRuntimeSetsAreEachUsed)$' -count=1 -parallel 4 -timeout 90s -json

All seven guards pass, with no guard implementation or pinned exemption edited. The carried SortedMap fields, registered native agreement programs and exact runtime witness remain intact. No new compiler-consuming package exists. The C mutant compilation now uses main's placeless native builder.

Vet command (vet.log, exit zero with no output):

    timeout 180 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/native/runtimeproducts ./internal/oracle ./stage1/cohere/json

Main compatibility command (main-compatibility.jsonl):

    timeout 180 go test ./internal/native -run '^(TestParallelMapCompilerABI|TestAsyncCompilerSuspensionsMatchNode|TestParallelCompilerSharedReferencesMatchNode)$' -count=1 -parallel 1 -timeout 90s -json

Flow corpus command (flow-corpus.jsonl and flow-corpus.stderr):

    timeout 720 go test ./internal/flow -run '^TestFlow' -count=1 -parallel 4 -timeout 660s -json

All 19 external source mutants were caught by the expected semantic assertion or sanitizer, with exit 1 and no build-failure kill. run-mutants.py records exact mutation sites and catchers; each patch is noncompilable evidence. mutants-restored.json verifies byte-for-byte restoration. The private runtime cache tests also execute their built-in allocation, identity, release-order, resurrection, canonical and leak-root mutants. The four array-runtime tests exercise the migrated native builder's private mutated archives.

| Source mutant | Catcher | Command wall s | Go leaf s |
|---|---|---:|---:|
| escape-Argument | TestV4EscapeArgumentRefusal | 16.843 | 4.75 |
| escape-Result | TestV4EscapeResultRefusal | 8.779 | 4.56 |
| escape-Passed | TestV4EscapePassedRefusal | 8.587 | 4.68 |
| escape-Returned | TestV4EscapeReturnedRefusal | 8.826 | 5.03 |
| escape-MixedProducer | TestV4EscapeMixedProducerRefusal | 6.669 | 3.73 |
| unsupported-domain | TestV4DirectRecursiveRefusal | 12.274 | 3.38 |
| unsupported-producer | TestV4DirectProducerRefusal | 17.363 | 6.08 |
| surface-borrow | TestV4EscapeAdapterName | 19.510 | 6.52 |
| bind-name-borrow | TestV4EscapeAdapterBindNameOwnership | 19.790 | 6.53 |
| identity-root | TestV4EscapeAdapterIdentity | 23.141 | 9.04 |
| metadata-inline | TestPortElementMetadataEmission | 6.059 | 1.24 |
| metadata-omitted | TestPortElementMetadataEmission | 5.458 | 1.42 |
| native-order | TestV4DirectCalleeOrder | 20.499 | 7.57 |
| javascript-order | TestV4DirectCalleeOrder | 77.999 | 5.87 |
| construct-witness | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 7.189 | 0.27 |
| boxed-witness | TestV4Lane5BoxedAdamicRefusal | 15.706 | 8.76 |
| factory-witness | TestV4Lane5FactoryWiderWriteRefusal | 15.458 | 9.03 |
| runtime-witness-mismatch | TestRuntimeSetsAreEachUsed | 23.284 | 2.66 |
| main-library-spread-omitted | TestLibrarySpreadRegExpUnion | 7.719 | 0.22 |

The javascript-order mutant's 77.999 s command includes rebuilding Go after a backend source edit; its actual test leaf is 5.87 s. No over-budget leaf was rerun. Observed catchers are refusal/path/fix assertions for escape/domain/producer/witness changes, ASan heap-use-after-free for both borrowing omissions, Node output comparison for identity and call-order changes, emitted-C certificate assertions for the metadata omissions, exact feature-set mismatch for the runtime witness, and the preserved main library-spread refusal assertion.

The known cold bridge TestProduct_ApiDriver from the earlier unit was not rerun. The reader package compiles and its cache-integrity control passes. The prior user-provided Loom cache record is separate from this unit's observations.

A fresh timing exposed one more V4-owned gate risk: the old TestOriginalArrayWitnessNode passed but took 69.442 s to run its 473 Node subprocesses. It was never rerun intact. Split it into 16 top-level TestOriginalArrayWitnessNode_0000 through _0015 functions, each calling t.Parallel() first, with a shared ordinal-modulo helper. The helper preserves the same four manifests, expected Node stdout and every probe source. It runs the assigned probes directly, without nested test groups. witness-shard-union.json names all 473 probes exactly once: nine shards have 30, seven have 29. No fixture was changed or deleted. Raw focused-results.jsonl preserves the over-budget observation; final-focused-results.json replaces that removed function with its 16 verified replacements.

The replacement shards all pass. witness-output-mutant.patch adds wrong bytes to the expected stdout in the shared helper: shard 0000 fails the same Node agreement assertion (witness-output-mutant.log). The mutant was restored. Post-split TestCallTargetReaders, naming, compiler census and parallel guards pass (post-split-guards.jsonl), and vet again exits zero with no output (post-split-vet.log).

| New top-level shard | Seconds |
|---|---:|
| TestOriginalArrayWitnessNode_0000 | 2.389 |
| TestOriginalArrayWitnessNode_0001 | 2.283 |
| TestOriginalArrayWitnessNode_0002 | 2.377 |
| TestOriginalArrayWitnessNode_0003 | 2.389 |
| TestOriginalArrayWitnessNode_0004 | 2.385 |
| TestOriginalArrayWitnessNode_0005 | 2.483 |
| TestOriginalArrayWitnessNode_0006 | 2.540 |
| TestOriginalArrayWitnessNode_0007 | 2.524 |
| TestOriginalArrayWitnessNode_0008 | 2.339 |
| TestOriginalArrayWitnessNode_0009 | 2.176 |
| TestOriginalArrayWitnessNode_0010 | 2.132 |
| TestOriginalArrayWitnessNode_0011 | 2.233 |
| TestOriginalArrayWitnessNode_0012 | 2.719 |
| TestOriginalArrayWitnessNode_0013 | 3.768 |
| TestOriginalArrayWitnessNode_0014 | 3.921 |
| TestOriginalArrayWitnessNode_0015 | 3.364 |

Only four implementation/test files differ from b7268d27: the two backend test-name declarations, the private array-runtime builder, and this Node witness test split. No package tests were weakened to make room for V4. The conservative assumption is that all 473 source-profile probes remain required, so they are partitioned rather than sampled.

Final validated inventory: {'focused': 314, 'focused_skips': ['TestViewAdapterPoolTSanPending', 'TestViewAdapterProgramUnderlyingPending', 'TestViewAdapterRegionAdoptionPending', 'TestViewAdapterRuntimeStaticsPending', 'TestV4EscapeAdapterBlameRuntimeCallback', 'TestV4Lane5FactoryReadonlyTargetControl', 'TestV4Lane5FactoryWiderWritePending'], 'focused_max': 34.892, 'flow_leaves': 859, 'flow_max': 52.02, 'flow_seconds': 125.755, 'reader_packages': 53, 'reader_tests': 48, 'reader_max': 33.844, 'external_mutants': 20}

Seven focused controls remain explicitly skipped for the prior documented pending runtime/readonly/factory coverage; these are listed in summary.json and their logs. No full package tests, full gate, ahra invocation, PR or Loom submission was run. All logs are retained, including setup mistakes and the single superseded over-budget aggregate; the final inventory has no over-budget leaf. All implementation Go files are gofmt-clean and source whitespace checks pass.
