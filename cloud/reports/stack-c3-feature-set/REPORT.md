Merged c79c7572 while preserving WASI test units and program-specific runtime feature selection.
Merge 2816d290; stack repair 3d1095b3 cherry-picked as da03166e.
Build, vet, stage3, scoped a-check, native/WASI matrices, split checker, tools and counts all pass.
Six branch mutants plus fuzzer default-runtime and WASI omitted-feature mutants are caught.
No full gate, new roadmap mechanism or unrelated package sweep was run.

Conflict: internal/native/wasm_test.go, three hunks. Preserve c79 global fixtures, shard ownership, compiler/runtime/link build caches and counted request units. Include each generated program's features in runtime flags and cache keys, then use that runtime's include directory and objects at the link. Keep the branch's three added feature fixtures. No dropped e3380f23 test split entered history. The stack-owned repair was cherry-picked after the merge without conflict.

Automatic-merge interaction: c79 production fuzzer runtime selection is already source-aware. The added test now holds successful matching native/backend output to Node, rather than expecting the obsolete shared-default-runtime failure. Default-runtime sharing for plain programs and genuine test262 mismatch rejection stay tested.

Commands/exits: feature-checks.json and compressed logs contain go build ./..., go vet ./internal/..., go test ./stage3/fixtures -count=1, TestRuntimeFeatureMismatch, TestRuntimeFeatureSets, TestSplitTSGoRuntimeFeatures, TestSplitTSGoAgrees, TestTSGoBuildSeesTheProgramsFeatures (with a real ADAMIC_CLANG_TSGO_ARCHIVE), TestNativeMutantUsesProgramsFeatures, TestFuzzerUsesProgramsFeatures, TestRunnerFeatureMismatchStopsAtLink and TestFuzzerSharesRuntimeLibrary. Scoped a-check compares against c79; zero changed .a files. Counts: GOMEMLIMIT=2GiB go test ./internal/oracle -run TestCountsAreRecorded -parallel=2 -count=1 -args -update-counts, exit 0, 170.356s. No added, removed, reordered or numerically changed count rows against either the merge baseline or c79.

WASI: ADAMIC_TEST_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH go test ./internal/native -run '^TestWASI$' -count=1 -timeout=20m -v. All 38 source-Node equivalent fixtures and requests pass. Explicit WASI_SYSROOT also runs matching/mismatch feature matrices on native and wasm32-wasi. Default setup had reset the optional sysroot variable; opt-in-skipped attempts are excluded as evidence.

Mutants and witnesses:
fixed-symbol: caught by mismatched runtime linked successfully; /tmp/adamic-feature-set-link-mutants/fixed-symbol/test.log
drop-reference: caught by mismatched runtime linked successfully; /tmp/adamic-feature-set-link-mutants/drop-reference/test.log
drop-retain: caught by mismatched runtime linked successfully; /tmp/adamic-feature-set-link-mutants/drop-retain/test.log
header-forces-runtime-features: caught by mismatched runtime linked successfully; /tmp/adamic-feature-set-link-mutants/header-forces-runtime-features/test.log
split-runtime-default-flags: unsanitized link rejection verified; /tmp/adamic-feature-set-link-mutants/split-runtime-default-flags/test.log
split-runtime-default-flags: ordinary matching-build test catches mutant; /tmp/adamic-feature-set-link-mutants/split-runtime-default-flags-positive-control/test.log
split-original-checker-default-flags: original checker fixture fails at unsanitized link; /tmp/adamic-feature-set-link-mutants/split-original-checker-default-flags/test.log

Additional mutants: restoring default fuzzer runtime selection fails TestFuzzerUsesProgramsFeatures at undefined adamic_runtime_features_closure_convention; omitting the runtime flags in retained WASI runtimeBuild fails arguments_length_extended.a with undefined adamic_runtime_features_closure_convention_closure_receivers. These are the intended link guards, not compiler-warning failures. Both negative runs return 1 with their named witness. See feature-fuzz-mutant.log.gz and feature-wasi-mutant.log.gz.

The repaired optional-field tests, including their existing presence/readiness/representation mutants, were also verified on the first stack member carrying the identical stack patch. No independent emitter patch was introduced. Setup uses GOPROXY=https://proxy.golang.org|direct; nproc=5, CPU quota 4. Build 39.518s, total 39.742s, with detailed lines in the first member's setup evidence. Go build cache was cleared after the sweeps because old mutation artifacts had consumed disk; no source or result evidence was deleted.
