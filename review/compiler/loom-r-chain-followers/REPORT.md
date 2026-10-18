Built: applied only f5a13b1's own compiler change to pinned main dbdf977f5b as compiler/loom-r-chain-followers.
Commits: source f5a13b10596bcd419badbd39f4fb229dbbdff203; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA in the final message.
Commands and outputs: all focused positive checks exit 0; exact argv, exits and command seconds in checks.json, all output in logs.
Mutants: all 22 from the source report caught: fifteen source overlays fail named tests, seven in-test output mutations disagree with Node in both backends.
Not covered: whole packages, the full Loom gate, WASI or ahra loom check; unchanged main count rows were preserved rather than remeasured.

Base and application: fetched origin and the exact source SHA, then created compiler/loom-r-chain-followers from dbdf977f5b8b0979d414ebe6f20e7aadd312e309. Used git cherry-pick --no-commit f5a13b10596bcd419badbd39f4fb229dbbdff203, so the applied change is source~1..source, never main..source. The base stays pinned to the requested SHA even if origin/main advances. The inherited source scope and its exclusions are unchanged. Existing consumer tests were not edited to fit this change. New evidence is under review/compiler/loom-r-chain-followers; older source reports are retained as historical evidence.

Conflict and resolution, one file and one hunk:

internal/oracle/counts.md, combined hunk @@ -427,19 -427,9 +427,25 @@ after array_literal_empty_union.a. Pinned main contributes thirteen async rows (async_coverage_discard/parameters/reject_eager/reject_empty/reject_nested/typeof/unions/values, async_nested/plain/three/throw/typeof). The source adds arrow_this_super/field/stored at the same insertion point. Saved conflicts.patch, then used git checkout --ours -- internal/oracle/counts.md, preserving pinned main's complete table without hand-merging any row. After the compiler and fixture checks passed, ran exactly one Linux -update-counts command with house clang 20.1.8, selecting the fifteen source fixture leaves as supported by counts_test.go's writeCounts. The canonical writer inserted the source rows in registration order, preserving the async rows and every other main row. No other file conflicted. The automatic merges retain pinned main's newer compiler behavior.

Counts: exactly fifteen added rows, no changed or removed existing row. All 1058 pinned-main rows keep their original values. TestCountsAreRecordedTable passed afterward. Counts refresh follows the command in counts.md's header and the partial-refresh preservation contract documented above writeCounts in internal/oracle/counts_test.go. The fifteen added rows are attributable one-for-one to this change's new registered oracle fixtures; the source and all fixtures compiled before the one refresh. counts-main.md.txt is the untouched base table, and counts-attribution.json independently compares every row.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/arrow_this_super.a | 5 | 5 | 8 | 8 | 5 | 0 |
| internal/oracle/testdata/arrow_this_field.a | 7 | 7 | 9 | 11 | 6 | 0 |
| internal/oracle/testdata/arrow_this_stored.a | 7 | 7 | 9 | 11 | 7 | 0 |
| internal/oracle/testdata/native-generic-empty-array.a | 3 | 3 | 1 | 5 | 3 | 0 |
| internal/oracle/testdata/generic_empty_array.a | 32 | 32 | 24 | 60 | 17 | 0 |
| internal/oracle/testdata/generic_function_value_arms.a | 18 | 18 | 72 | 99 | 7 | 0 |
| internal/oracle/testdata/generic_function_value.a | 15 | 15 | 27 | 45 | 7 | 0 |
| internal/oracle/testdata/phantom_brands.a | 15 | 15 | 29 | 42 | 6 | 0 |
| internal/oracle/testdata/phantom_undefined.a | 1 | 0 | 5 | 4 | 1 | 0 |
| internal/oracle/testdata/phantom_catch.a | 8 | 8 | 10 | 18 | 2 | 0 |
| internal/oracle/testdata/host_phantom_overload_primitive.a | 25 | 25 | 13 | 43 | 7 | 0 |
| internal/oracle/testdata/host_phantom_optional_array.a | 5 | 5 | 2 | 7 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | 2 | 2 | 3 | 6 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/number_spellings.a | 85 | 85 | 2 | 88 | 3 | 0 |
| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | 33 | 33 | 2 | 38 | 3 | 0 |

Runtime C lines: NONE. git diff source~1 source -- internal/native/runtime is empty, as is git diff base HEAD/worktree -- internal/native/runtime. Thus the runtime +/- lines are identical to the source's own runtime lines, with no line needing runtime clearance. runtime-comparison.json records both diffs and an empty changed_lines array. Code generation changes live in Go, as in the source; main's runtime C remains intact.

Verification and reader coverage:

- lower-focus: exit 0, 10.346 command seconds.
- oracle-focus: exit 0, 28.462 command seconds.
- ir-focus: exit 0, 13.699 command seconds.
- fixtures: exit 0, 8.002 command seconds.
- flow-consumers: exit 0, 20.137 command seconds.
- fresh-consumers: exit 0, 7.700 command seconds.
- native-consumers: exit 0, 11.956 command seconds.
- javascript-consumers: exit 0, 1.423 command seconds.
- stage3: exit 0, 19.935 command seconds.
- guards-ir: exit 0, 2.532 command seconds.
- guards-runtime: exit 0, 3.745 command seconds.
- guards-registration: exit 0, 2.643 command seconds.
- guards-parallel: exit 0, 1.779 command seconds.
- counts-regeneration: exit 0, 6.271 command seconds.
- counts-table: exit 0, 2.575 command seconds.
- readers: exit 0, 9.434 command seconds.
- vet: exit 0, 4.154 command seconds.

All 15 exact source oracle fixture leaves passed uncached against source Node, JavaScript emission, native sanitizers and leak checks. All 12 added top-level fixtures also passed the flow and freshness corpus remainders; flow's 27-program remainder runs single assignment, mutation ranges, graph paths and liveness, retaining its independent Node observations, and finished in 4.500 test seconds. Freshness's remainder passed in 1.450 seconds. The three nested scanner fixtures are held by their oracle leaves. Every new top-level fixture lowers; no newly added refused top-level fixture needed relocation to a *_refused directory. No consumer or corpus list was changed. Source lowering tests, native memory-plan consumers, a JavaScript representation consumer, the seven changed stage3 status rows, and IR readers passed. Setup's authorized go build also compiled all repository packages, retaining main's product key behavior.

TestCallTargetReaders used -timeout 90s. go vet passed on internal/flow, fresh, ir, javascript, lower, native and oracle. Guards also passed for SortedMap-only IR program maps, runtime-set witnesses, native agreement registration, and test parallelism. The new node holds no map, no runtime feature set was added, no package was added, and no compiler-consumer declaration is needed. Products and tests introduce no build-directory or machine-path literals. Changed Go files are gofmt clean. No code was copied from cohere; the source retains its pinned checker/shim inference reference.

Each test's seconds, including parent tests, leaves and intentional mutant failures, are in test-times.tsv and test-times.json. Maximum measured test: 19.160 seconds. All runs use GOMAXPROCS=4, -parallel 4, -count=1, -timeout 90s and ADAMIC_GATE_UNCACHED=1 on one four-CPU quota instance. No test crossed 60 seconds, so none was deleted, split or rerun after crossing the limit. Compilation/startup time is recorded separately as command wall seconds. Each test command has its Go timeout and a 240-second subprocess hard limit; the runner has a 900-second limit and was monitored. Test stdout/stderr went directly to log files and was never piped.

Mutants and catchers:

- non-null-check (source overlay): TestCheckedNonNullTypeScript/undefined (0.040s), TestCheckedNonNullTypeScript (0.000s).
- non-null-narrowing (source overlay): TestCheckedNonNullTypeScript/narrowed (0.040s), TestCheckedNonNullTypeScript (0.000s).
- nested-empty (source overlay): TestNestedEmptyArrayElementKinds/strings (0.270s), TestNestedEmptyArrayElementKinds/objects (0.180s), TestNestedEmptyArrayElementKinds (0.000s).
- string-concatenation (source overlay): TestNativeAgreesWithNode_081/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a (0.030s), TestNativeAgreesWithNode_081 (0.000s).
- phantom-primitive (source overlay): TestPhantomCastsAreErased (0.040s).
- proven-predicates (source overlay): TestCensusComputedRelation (0.090s).
- generic-empty (source overlay): TestNativeAgreesWithNode_059/internal/oracle/testdata/generic_empty_array.a (0.040s), TestNativeAgreesWithNode_059 (0.010s).
- debugger (source overlay): TestDebuggerNativeEmitsNothing/.a (0.050s), TestDebuggerNativeEmitsNothing/.ts (0.050s), TestDebuggerNativeEmitsNothing (0.000s).
- generic-values (source overlay): TestGenericFunctionValuesHaveSeparateInstances (0.030s).
- generic-namespace-interaction (source overlay): TestGenericFunctionValueNamespaceDirectCallsKeepIdentity (0.030s).
- phantom-overload-interaction (source overlay): TestPhantomOverloadOnlyBrandsKeepMainProof (0.030s).
- proven-cleanup (source overlay): TestCensusGenericRefusalInsideClosure (0.030s).
- arrow-constructor-frame (source overlay): TestArrowFieldCaptureFrames (0.130s).
- arrow-receiver-capture (source overlay): TestArrowSuperCaptureFrames (0.050s), TestArrowStoredCaptureFrames (0.050s).
- phantom-fresh (source overlay): TestFreshCorpusRemainder/../oracle/testdata/host_phantom_overload_primitive.a (0.030s), TestFreshCorpusRemainder/../oracle/testdata/phantom_brands.a (0.030s), TestFreshCorpusRemainder/../oracle/testdata/phantom_catch.a (0.040s), TestFreshCorpusRemainder/../oracle/testdata/phantom_undefined.a (0.030s), TestFreshCorpusRemainder (1.400s).
- empty-array-falsiness (IR output mutant): TestEmptyArrayRuntimeMutants/empty_array_is_falsy (0.600s).
- fresh-array-sharing (IR output mutant): TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared (0.640s).
- generic-wrong-result (IR output mutant): TestGenericFunctionValueWrongResultMutant (18.910s).
- numeric-concatenation-spelling (IR output mutant): TestScannerConcatenationMutant (18.970s).
- boolean-concatenation-spelling (IR output mutant): TestScannerPrimitiveConcatenationMutants/boolean (0.550s).
- undefined-concatenation-spelling (IR output mutant): TestScannerPrimitiveConcatenationMutants/undefined (0.510s).
- null-concatenation-spelling (IR output mutant): TestScannerPrimitiveConcatenationMutants/null (0.580s).

The fifteenth source mutant is phantom-fresh from the latest source REPORT.md: removing ir.PhantomMember from the existing conservative opaque-call freshness model is caught by TestFreshCorpusRemainder. Every stored source mutant is reconstructed as only its mutation delta against f5a13b1, then applied to the rebased file; an old whole-file snapshot never replaces main's newer code. Each source run exited 1 with named test failures, not a Go build failure. Arrow mutants are caught by IR frame assertions before C emission. The seven in-test IR mutations require stdout-only disagreement with Node in both backends; build/sanitizer/exit/leak failures do not qualify. Exact failure messages and catchers are retained in mutants/results.json and per-mutant logs.

Toolchain and setup: exported GOPROXY=https://proxy.golang.org|direct; ran timeout 600 bash cloud/setup.sh and sourced its printed /workspace/adamic-tools/env.sh. Setup succeeded and proved the sanitizer with a real heap-buffer-overflow probe. nproc=5, cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node v24.19.0. Exact timing lines:

- setup: go ready (0.088s)
- setup: node ready (0.096s)
- setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.013s
- setup: markdown dependencies ready (0.259s)
- setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.522s)
- setup: submodules ready (1.390s)
- setup: shared cache ready (2.502s)
- setup: go build ready (65.588s)
- setup: test binaries deferred (use --warm-tests) (65.667s)
- setup: build cache warm (65.668s)
- setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.08 0.03 0.01 1/137 918 load-after=4.75 1.27 0.43 1/152 2050
- setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (65.700s)
- setup: source /workspace/adamic-tools/env.sh
- setup: logs /tmp/adamic-gate/setup.mnoE5T

Scope assumptions: the named source commit's own diff is authoritative, including its conservative opaque-call freshness treatment and parallel source tests. No freshness precision is inferred beyond that model. The pinned base overrides the generic current-main instruction. The full gate belongs to Loom, so checks are focused rather than whole-package runs. No ahra CLI, PR creation or Loom submission was run. The old branches remain untouched and the new branch is pushed without force. Historical evidence bytes and inherited fixture whitespace are preserved.
