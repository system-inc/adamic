Built source-aware lint profiling toward lowering chain #wj4pmt1 and feature link check #hmab710.
Code commit: 3535387095d50f05d80e2717aed468843e59bc18; base: 2a28375f9c8c487d29abb89f38489858fc9d5c92.
Both requested product leaves, all profile products, Node comparison, runtime-library tests and lane checks pass.
Two feature-selection reversion mutants fail at the exact undefined closure-convention symbol.
Cold preparation exceeds the pool ceiling; opt-in full artifact corpora and WASI execution were not covered.

The unit-specific base instruction was used: compiler/chain-lint-features starts at origin/compiler/optional-presence-next, 2a28375f. No other worker branch was merged. PR 288 was read only to match its SourceFlags export. An exact text comparison confirmed the function, doc comment and placement match.

Lint's build helper stays local to its package. profile_compilation_main_test.go line 194 uses SourceFlags for the cache recipe; line 205 delegates the profiled build; lines 357-374 hold the shared linker, with RuntimeLibraryForSource at 358 and SourceFlags at 366. profile_test.go lines 82-84 reuse it. This removes duplicate runtime copying and linking code. The scanner and counted variants already call native.Build, which selects source features.

Other paths fixed outside PR 288's fuzz, test262 and decode-ascii scope:

| File | Path and changed lines |
| --- | --- |
| stage1/typescript/scanner/profile_test.go | Profile artifacts, 68 and 72 |
| stage1/cohere/css/profile_test.go | Profile artifacts, 106 and 110 |
| stage1/cohere/markdownblocks/malformed_events_independent_test.go | Native emitted-C builder, 339; runtime already source-aware |
| stage1/cohere/markdownblocks/list_layout_shards_test.go | Normal builder, 396; canary flags/runtime, 475 and 482 |
| stage1/cohere/markdownblocks/support_test.go | Native mutant builder, 342; runtime already source-aware |
| internal/native/units_measure_test.go | Trace linker and prewarm, 59, 68 and 152; flag metadata at 100 |
| internal/native/parser_construction_test.go | Custom runtime mutant builder, 143; SourceFlags also reaches cachedRuntime, preserving the mutated snapshot |
| stage3/drivers/scanner/build-metrics.go | Actual feature-aware runtime prewarm and flag metadata, 47 and 51 |
| internal/oracle/wasi_test.go | Emitted object compilation, 178; no runtime link in this path |

Remaining plain calls outside PR 288 are intentional feature-mismatch controls, hand-written C probes, or runtime identity checks without a program. Most remaining native.Flags calls describe cache recipes around native.Build. ESTree's manual builder already propagates the emitted features.

The new TestProfileRuntimeFeaturesMatchNode reuses internal/oracle/testdata/arguments_length_value_count.a and asserts that its emitted C enables ADAMIC_CLOSURE_CONVENTION. It calls the actual shared profile linker, runs source on Node, and compares native stdout. Both print "77\n2\n3\n". No new fixture or counts row was added; counts.md therefore needed no refresh.

Final observed test leaf durations:

| Leaf | Seconds | Result |
| --- | ---: | --- |
| TestProfileRuntimeFeaturesMatchNode | 7.22 | pass |
| TestProduct_ProfileCompilationProfiled | 14.68 | pass, run alone |
| TestProduct_ProfileCompilationOracleOutputs | 2.40 | pass, run alone |
| TestProduct_ProfileCompilationScanner | 2.02 | pass |
| TestProduct_ProfileCompilationCounted | 1.35 | pass |
| TestProduct_ProfileCompilationLowered | 1.03 | pass |
| TestProduct_ProfileCompilationC | 1.05 | pass |
| TestProduct_ProfileCompilationJavaScript | 1.08 | pass |
| TestProduct_ProfileCompilationOracle | 4.71 | pass |
| TestProfileCompilation_000 | 3.12 | pass, full profile output comparison |
| TestRuntimeKeyIncludesEveryInput | 0.41 | pass |
| TestRuntimeKeyKeepsBoundaries | 0.00 | pass |
| TestRuntimeCacheKeepsCountFlags | 11.49 | pass |
| TestRuntimeCacheRebuildsChangedSources | 0.70 | pass |
| TestRuntimeCacheConcurrentBuilders | 2.56 | pass |
| TestRuntimeCacheConcurrentProcesses | 0.39 | pass |
| TestProduct_MarkdownMalformedEventsNative | 38.83 | pass |
| TestNativeMutantUsesProgramsFeatures | 6.49 | pass |
| TestMeasureClangUnits, small closure fixture | 10.98 | pass, isolated warm rerun |
| TestParserConstructionRuntimeMutantArrayExtrasInJSON | 6.31 | pass, mutant caught by Node disagreement |
| TestParserConstructionRuntimeMutantDeclarationOrder | 6.52 | pass, mutant caught by Node disagreement |
| TestParserConstructionRuntimeMutantSynthesizedKey | 6.63 | pass, mutant caught by Node disagreement |
| TestCallTargetReaders | 2.41 | pass after final edits |

Every ordinary test command used source /workspace/adamic-tools/env.sh, go test -count=1 -v -timeout 90s and a hard outer timeout. Tests wrote stdout/stderr to the matching .log file. Product leaves used individual anchored filters: ^TestProduct_ProfileCompilation<Leaf>$. Other completed filters were ^TestProfileRuntimeFeaturesMatchNode$, ^TestProfileCompilation_000$, ^TestRuntime(Key|Cache), ^TestProduct_MarkdownMalformedEventsNative$, ^TestNativeMutantUsesProgramsFeatures$, ^TestParserConstructionRuntimeMutant and ^TestCallTargetReaders$. No whole-package test run or full gate was run.

TestMeasureClangUnits used ADAMIC_CLANG_MEASURE=/workspace/adamic/review/compiler/chain-lint-features/clang-measure and ADAMIC_CLANG_PROGRAM=lint-harness. Its input was emitted C for the existing arguments_length_value_count.a fixture; the changed source replaced the single 0x1.34p+06 with 0x1.38p+06, changing exactly one of three split units. The first cold, overlapping run passed in 62.42s; the isolated warm run passed in 10.98s. The full external profiling corpus was not exercised.

Mutants:

- profile-runtime-mutant.go.txt reverts RuntimeLibraryForSource and SourceFlags only in compilationProfileBuild. go test ./stage1/cohere/lint -overlay=review/compiler/chain-lint-features/mutant-overlay.json -run '^TestProfileRuntimeFeaturesMatchNode$' -count=1 -v -timeout 90s fails in 5.31s. The linker reports undefined reference to adamic_runtime_features_closure_convention from .data.rel.ro.adamic_runtime_feature_reference.
- measure-mutant.go.txt reverts the timing harness's runtime selections to RuntimeLibrary. The same measurement environment and go test ./internal/native -overlay=review/compiler/chain-lint-features/measure-overlay.json -run '^TestMeasureClangUnits$' -count=1 -v -timeout 90s fail in 8.82s at the same undefined feature symbol.
- The three existing parser-construction mutants were run and caught by their existing Node comparisons. Their exact mutations and observed different outputs are recorded in parser-construction.log.

The new link mutants fail at linking, not from compiler warnings or sanitizer findings. Live sources were not mutated: Go overlays read .go.txt evidence.

Additional commands and observations:

- timeout 240 bash cloud/setup.sh with GOPROXY='https://proxy.golang.org|direct': exit 124 during build-cache warming. Timing lines: node ready 0.028s; go ready 0.031s; clang ready 0.235s; markdown dependency installation step 1.255s and ready 1.348s; submodules ready 23.616s; shared cache ready 30.567s. The generated /workspace/adamic-tools/env.sh was sourced for subsequent commands. nproc=5; cpu.max=400000 100000, a four-CPU quota.
- Initial outer 120s cold Go compilation attempts expired before test bodies. Concurrent compilation exhausted /tmp during two product builds. Only identified scratch directories from completed attempts were removed, restoring space.
- Cold source emission exceeded the 90s pool ceiling. The isolated command timeout -k 5 360 go test ./stage1/cohere/lint -run '^TestProduct_ProfileCompilationC$' -count=1 -v -timeout 300s passed in 91.95s, with 90.54s in C emission. Cold 90s scanner/profile attempts were killed inside emission. OracleOutputs also hit 90s while compiling its Go oracle. Logs retain those failures; all final product runs passed after preparation.
- go run ./cmd/adamic c internal/oracle/testdata/arguments_length_value_count.a generated metrics-source.c. go run stage3/drivers/scanner/build-metrics.go review/compiler/chain-lint-features/metrics-source.c review/compiler/chain-lint-features/scanner-metrics 1 passed. The report includes -DADAMIC_CLOSURE_CONVENTION=1. Unsplit, split-cold and split-warm binaries all matched source Node byte for byte.
- go vet ./internal/native ./internal/oracle ./stage1/cohere/lint ./stage1/cohere/markdownblocks ./stage1/cohere/css ./stage1/typescript/scanner: exit 0, no findings.
- go test ./stage1/cohere/css ./stage1/typescript/scanner -run '^(TestCSSProfileArtifacts|TestProfileArtifacts)$' -count=1 -v -timeout 90s: packages compile; both leaves skip because their opt-in artifact directories are unset.
- git diff --check: exit 0.
- The mandated lane command was run after the code commit. The checkout initially tracked only main, so the fast-gate and merge-tree refs were first fetched with explicit remote-tracking refspecs. Final mandated command: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -. Exit 0: lane checks 19.0 s: gofmt and tools on 302 Go files, t.Parallel on 24 test packages; a-check 103 .a files; vet 24 packages. This includes the inherited chain diff against main.

Limits: no full gate, WASI execution, complete markdown list corpus, or opt-in CSS/scanner/lint artifact corpus. PR 288's three paths remain its dependency, with an identical export to avoid a merge conflict. The observed cold build ceilings remain a separate build-grain issue; this unit fixes feature-consistent linking and proves the profile link regression.
