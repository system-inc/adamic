# Runtime feature sets must agree at link

Roadmap step 04 follow-up, task #hmab710; remaining tool paths #qbmxqke.
Base: origin/main 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8, including split-build-flags e96b7d43.

Every translation unit that includes adamic.h retains a relocation to a readable
symbol computed from the five feature macros. features.c defines only the symbol
for the runtime's own set. Compiler optimization and linker section removal cannot
remove the reference. There is no executable runtime check. The runtime cache
keeps the header unchanged: inserting runtime feature defines into that header
would make a program requesting fewer features silently adopt the runtime's set.

The feature set uses macro definedness, matching the runtime's existing #ifdef
layout decisions. ADAMIC_COUNT and ADAMIC_TSGO are outside the five-feature
contract in library.go. No emitter, lowering or oracle registration changed.

Observed without sanitizers, for a closure-convention program against
RuntimeLibrary(""):

```
undefined reference to `adamic_runtime_features_closure_convention'
```

The original TestSplitTSGoAgrees fixture with the third split-build-flags mutant
also stops at this symbol, without sanitizers. Its split runtime uses Flags(options)
instead of sourceFlags(source, options); the test's options are changed only in a
Go overlay to remove sanitizers. Neither mismatched program executes.

## Build paths

| Path | Runtime selection and result |
| --- | --- |
| Native Build | RuntimeLibraryForSource selects the program's features; all 784 registered native oracle fixtures agree with Node. |
| Options.Split / buildUnits | Existing sourceFlags applies the program's features; all 784 fixtures agree with Node with ADAMIC_NATIVE_SPLIT=1. |
| BuildTSGo | Existing sourceFlags preserves features before the checker header; TestTSGoBuildSeesTheProgramsFeatures passes with a real checker archive. |
| BuildSplitTSGo | Existing sourceFlags selects both unit and runtime features; TestSplitTSGoAgrees passes cached and uncached with the archive. An additional unsanitized split-checker witness agrees with Node. |
| Stage 1 Markdown mutant helper | Changed from RuntimeLibrary to RuntimeLibraryForSource using the same emitted C that it compiles; its new feature-bearing Node comparison passes under sanitizers. |
| wasm32-wasi oracle helper | Changed from one shared object set to cached object sets selected by featureFlags; 38/38 Node comparisons and the counted request probe pass. |
| internal/fuzz | Keeps its shared default runtime. A feature-bearing program now fails at link; the plain-runtime control still passes. |
| cmd/adamic-test262 | Keeps its shared default runtime. A feature-bearing program is reported as a clang failure at link. |
| cloud/reports/decode-ascii | Keeps its shared default runtime. The actual build helper fails at link for arguments_length_extended.a, naming closure_convention_closure_receivers; it never executes the binary. |

The guard exposed the WASI helper's previous mismatch on closures_throw.a:
the program enabled the closure convention and the helper compiled the runtime
without it. It previously linked. The helper now builds the matching runtime.
Three existing feature-bearing oracle fixtures were added to the WASI test list;
no fixture source, expected output, stage 3 observation or count entry changed.

## Local checks

All output is stored in log files. Commands source /workspace/adamic-tools/env.sh.
The Node comparisons use ADAMIC_GATE_UNCACHED=1. The checker archive is built with:

```
go build -buildmode=c-archive -o /tmp/feature-link-tsgo.a ./bridge/tsgo/archive
npm ci --prefix stage3/api
```

On unchanged main, before editing:

```
go test ./internal/native ./internal/oracle ./stage3/fixtures -count=1 -timeout 30m
```

All passed: native 477.477s, full oracle 436.456s, stage 3 74.496s.
Log: /tmp/feature-link-main-baseline.log.
The checker-archive tests separately passed (98.995s;
/tmp/feature-link-main-checker.log). The original WASI suite separately passed
35/35 fixtures plus requests (64.062s; /tmp/feature-link-main-wasi.log).

On the changed branch:

| Command | Result and log |
| --- | --- |
| go test ./internal/native -run '^TestRuntimeFeatureMismatch$' -count=1 -v | Both mismatch directions rejected at link; matching programs run. /tmp/feature-link-mismatch-final.log |
| go test ./internal/native -run '^TestRuntimeFeatureSets$' -count=1 -v | Native and WASI: each has 32 matching sets, 160 one-bit mismatches, and one unused mixed unit; all pass with section removal. /tmp/feature-link-matrix-final.log |
| ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/feature-link-tsgo.a go test ./internal/native -run '^TestSplitTSGoRuntimeFeatures$\|^TestSplitTSGoAgrees$\|^TestTSGoBuildSeesTheProgramsFeatures$' -count=1 -timeout 30m -v | Pass, 48.527s. /tmp/feature-link-checker-final.log |
| ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASI$' -count=1 -timeout 30m -v | WASI SDK clang on PATH, WASI_SYSROOT set; 38/38 plus requests, 48.523s. /tmp/feature-link-wasi-final.log |
| go test ./internal/oracle -run '^TestNativeAgreesWithNode$\|^TestCountsAreRecorded$' -count=1 -timeout 30m -v | All 784 fixtures and recorded counts pass, 395.944s. /tmp/feature-link-oracle-native.log |
| ADAMIC_NATIVE_SPLIT=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$' -count=1 -timeout 30m -v | All 784 fixtures pass, 386.752s. /tmp/feature-link-oracle-split.log |
| go test ./stage3/fixtures -count=1 -timeout 30m -v | Pass, 87.485s. /tmp/feature-link-stage3-final.log |
| go test ./stage1/cohere/markdownblocks -run '^TestNativeMutantUsesProgramsFeatures$' -count=1 -timeout 10m -v | Node comparison passes, 0.953s. /tmp/feature-link-stage1-helper.log |
| go test ./internal/fuzz ./cmd/adamic-test262 -run '^TestFuzzerFeatureMismatchStopsAtLink$\|^TestRunnerFeatureMismatchStopsAtLink$\|^TestFuzzerSharesRuntimeLibrary$' -count=1 -timeout 10m -v | Link-stop tests and default-runtime control pass. /tmp/feature-link-tools-final.log |
| go run cloud/reports/decode-ascii/build.go internal/native/runtime internal/oracle/testdata/arguments_length_extended.a native /tmp/feature-link-decode-mismatch | Expected link failure, wrapper exit 1, wanted symbol closure_convention_closure_receivers. /tmp/feature-link-decode-mismatch.log |
| go vet ./internal/native ./internal/fuzz ./cmd/adamic-test262 ./stage1/cohere/markdownblocks | Pass. /tmp/feature-link-vet-final.log |

Table command alternatives use a vertical bar inside the test regexp; run them
as the quoted regular expressions shown. All modified Go files were gofmt'd,
and git diff --check passes. The full changed native package is deferred to the
gate with the commit trailer Gate-runs: deferred. No new registered fixture
requires count regeneration; TestCountsAreRecorded confirms the current file.

## Mutants actually run

The committed runner uses Go overlays and isolated runtime caches, leaving the
checkout unchanged:

```
ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/feature-link-tsgo.a python3 internal/native/testdata/run-feature-set-link-mutants.py
```

| Mutation | Observation |
| --- | --- |
| Fixed symbol for every set | The mismatch links successfully; TestRuntimeFeatureMismatch fails. |
| Remove the header reference | The mismatch links successfully; the same test fails. |
| Remove retain, keeping used | Linker section removal discards the reference; the mismatch links and the test fails. |
| Restore runtime feature defines in the cached header | The program requesting no features silently adopts the closure convention; the reverse mismatch links and the test fails. |
| splitTSGoRuntime uses Flags(options) | The dedicated unsanitized negative witness passes by observing the named undefined symbol; its ordinary Node-comparison mode fails at link. |
| Same third mutant on the original checker fixture | TestSplitTSGoAgrees fails at the named undefined symbol with sanitizers removed through a second overlay. |

Guard and dedicated split logs: /tmp/feature-link-mutants.log.
Original checker log: /tmp/feature-link-original-checker-mutant.log.
Individual overlay, source and test logs remain under
/tmp/adamic-feature-set-link-mutants/.

## Limits and toolchain

Linux native and wasm32-wasi were linked and exercised. A minimal header-reference
probe compiled for arm64-apple-macos14, but no macOS executable was linked or run.
The changed full native package and complete stage 1 Markdown package are deferred;
only the named stage 1 helper witness ran. The three shared-runtime tools reject
feature-bearing inputs; they do not yet rebuild a runtime per feature set.

GOPROXY=https://proxy.golang.org|direct was set before setup. Both setup invocations
passed; nproc is 5 (cpu.max 400000 100000). Versions: Go 1.27.1, Node 24.19.0,
clang 20.1.8, WASI SDK 27 (clang 20.1.8-wasi-sdk). Setup timing lines follow.

```
/tmp/feature-link-setup.log
setup: go ready (0.021s)
setup: node ready (0.019s)
setup: submodules ready (0.066s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: markdown dependencies ready (0.069s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.181s)
setup: go build ready (45.667s)
setup: test binaries deferred (use --warm-tests) (45.876s)
setup: build cache warm (45.877s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (45.906s)
/tmp/feature-link-wasi-setup.log
setup: go ready (0.021s)
setup: node ready (0.024s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.006s
setup: submodules ready (0.072s)
setup: markdown dependencies ready (0.074s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.167s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk) (14.708s)
setup: go build ready (47.912s)
setup: test binaries deferred (use --warm-tests) (48.580s)
setup: build cache warm (48.588s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (48.704s)
```
