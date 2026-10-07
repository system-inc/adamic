# Gate input provisioning and first main audit

Built --gate-inputs: seven locked npm directories, pinned TypeScript, deterministic 100 MiB file and checker archive, outside /root and exported by env.sh.
Implementation 156e982f1b1a51421612060c4a6873cf0b1e5e82 on devtools/gate-inputs, from developer-tools bf35c1cbf0cfba856abb081d094da7a1946c8e4d.
Fetched main 39638d9e278d38bb5aeae887f46d55a70e47aaad: 16 SKIP without the flag, 16 PASS uncached with it; the seventeenth test is absent.
All 29 cache, integrity and generated-header mutants caught; all ten actual installed trees byte-identical after uncached preparation.
No product fixes or main pushes; no whole gate, ARM or fresh VM run; absent-main split test separately skips and passes on developer-tools.

## Current main proof

The user cited historical main 2e16546. Main fetched at the start of this unit is 39638d9e278d38bb5aeae887f46d55a70e47aaad. It was checked out with git worktree add --detach /workspace/gate-inputs-main origin/main. No merges, cherry-picks, copied branch source or source overlays were applied to main. All changes committed by this unit are under cloud/.

```
ADAMIC_SETUP_REPOSITORY=/workspace/gate-inputs-main bash /workspace/adamic/cloud/setup.sh
source /workspace/adamic-tools/env.sh
python3 /workspace/adamic/cloud/run-gate-input-tests.py /workspace/gate-inputs-main /tmp/gate-inputs-without without
ADAMIC_SETUP_REPOSITORY=/workspace/gate-inputs-main bash /workspace/adamic/cloud/setup.sh --gate-inputs
source /workspace/adamic-tools/env.sh
python3 /workspace/adamic/cloud/run-gate-input-tests.py /workspace/gate-inputs-main /tmp/gate-inputs-with with
```

The runner sets ADAMIC_GATE_UNCACHED=1. Each row runs go test ./<package> -run <anchored exact test/subtest> -count=1 -timeout=30m -json, with two concurrent checks. Test output is written directly to JSON event logs and readable logs are extracted afterwards. Exact commands, full build flags, CPU quota, cache state, loads before/after, inputs and durations are in each row JSON. Requested subtests retain required parent corpus setup and their own child mutants.

| Package | Test | Without flag | With flag on main | First differing line |
| --- | --- | --- | --- | --- |
| stage1/cohere/lint | `TestCompilerAndStage1Agree` | SKIP | PASS | none |
| stage1/typescript/parser | `TestCompilerExpressionsAgree` | SKIP | PASS | none |
| stage1/typescript/parser | `TestWholeCompilerAgrees` | SKIP | PASS | none |
| stage1/cohere/css | `TestThePortParsesAsGoCohereDoes/PostCSS` | SKIP | PASS | none |
| stage1/cohere/graphql | `TestThePortParsesAsGoCohereDoes/as_graphql-js` | SKIP | PASS | none |
| stage1/cohere/mediaquery | `TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser` | SKIP | PASS | none |
| stage1/cohere/selector | `TestThePortParsesAsGoCohereDoes/as_postcss-selector-parser` | SKIP | PASS | none |
| stage1/cohere/selector | `TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars` | SKIP | PASS | none |
| stage1/cohere/values | `TestThePortParsesAsGoCohereDoes/as_postcss-values-parser` | SKIP | PASS | none |
| stage1/cohere/json | `TestUpstreamNumericSeparatorGap` | SKIP | PASS | none |
| stage1/cohere/json | `TestUpstreamRepositoryCorpusParity` | SKIP | PASS | none |
| stage1/cohere/json | `TestExternalComparisonCatchesThreePrinterMutants` | SKIP | PASS | none |
| stage1/cohere/css | `TestCSSPrinterAgreesWithGo/default` | SKIP | PASS | none |
| stage1/cohere/css | `TestCSSPrinterAgreesWithGo/narrow` | SKIP | PASS | none |
| stage1/cohere/css | `TestCSSPrinterBoundaryProofs` | SKIP | PASS | none |
| stage1/cohere/gitignore | `TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower` | SKIP | PASS | none |
| internal/native | `TestSplitTSGoAgrees` | ABSENT | ABSENT | not applicable: absent test |

There were no new assertion failures among the 16 checks present, so there is no failing first differing line. PASS preserves explicitly recorded differences: JSON reports its nine known upstream differences, with full answers retained in with/json-corpus-upstream-differences.txt. CSS default and narrow each report 4,463 identical npm formats, 19,013 shared refusals and 20 recorded discrepancy occurrences (BOM 10, CR 2, whitespace 2, yaml 6). Lint compares 303 compiler/stage1 files and 20,440,938 identical Go/Node/native bytes.

TestSplitTSGoAgrees exists on developer-tools but not this main. Main filtered go test reports no tests to run, recorded ABSENT. No test was copied to main to turn that into a pass. Supplemental branch proof sets ADAMIC_CLANG_TSGO_ARCHIVE empty for SKIP, then uses a separate branch archive for PASS, cached/uncached split builds each matching 86 whole-checker bytes. See gate-inputs-split-branch-{without,with}.log. This is not a seventeenth main PASS.

## Pins and caching

See ../../gate-inputs/README.md for all exports, expected paths and pin authorities. Seven npm directories each have a checked-in lockfile and npm ci. npm 11.9.0 bootstrap SHA512 and package lock integrity are checked; scripts are disabled; manifest copies are immutable during installation. CSS uses postcss 8.5.16 and postcss-scss 4.0.9. Selector imports postcss for corpus extraction without a version assertion; its REPORT.md used 8.5.16, selected here. Other package versions have explicit test/driver pins. JSON and CSS printers use separate locked Prettier 3.9.6 directories.

TypeScript is a full depth-one checkout at exactly 050880ce59e30b356b686bd3144efe24f875ebc8, verified by HEAD and Git fsck. No sparse source subset is substituted. Its warm stamp checks URL, pin, helper and actual checkout bytes/modes/HEAD. The deterministic file is exactly 104,857,600 bytes, using the test recipe big-newline-#, x bytes, final newline, checked by size and SHA256. ADAMIC_GITIGNORE_LARGEST actually reads only nonempty opt-in, not a file path; exporting the installed path enables its own equivalent 100 MiB generation. The clean-main repeat catches the size-limit mutant at line 106 on both native and Node.

The checker build is go build -trimpath -buildvcs=false -buildmode=c-archive -o <outside-root>/tsgo.a ./bridge/tsgo/archive. Go validates the source/dependency action closure in exe mode, because go list in c-archive mode emits a stray .h on cold actions. Both validation and link flags are keyed, along with HEAD, helper, Go version/environment, C compiler version and dependency build IDs. Dirty source causes a rebuild. Archive/header filenames, bytes and modes are validated on warm hits; inputs are checked again before publishing a completed stamp.

ADAMIC_GATE_UNCACHED=1 bypasses all new stamps. All ten actual installed trees were captured, prepared again uncached and compared by filename/content/mode digest; gate-inputs-real-uncached.json records equal before/after hashes. Independent real npm, local Git and C archive tests exercise damage repair, pin changes, manifest edits, bad integrity and dirty source.

## Measurements

Three fresh input directories with empty npm download caches were interleaved with three warm checks, alternating order on the same box and same main commit. Instrument: python3 cloud/measure-gate-inputs.py /workspace/gate-inputs-main /tmp/gate-inputs-timings-final. Exact spawned helper commands and all flags/loads appear in timings/timings.json and individual logs. These are sequential phase measurements, not a sum claimed as parallel setup wall time.

| Loop | Phase | Fresh install | Warm check | Instrument |
| --- | --- | ---: | ---: | --- |
| 1 | Seven npm directories | 5.681s | 0.565s | helper phase npm, exact command in timings.json |
| 1 | TypeScript and 100 MiB file | 18.226s | 18.522s | helper phase corpora |
| 1 | Checker archive | 176.479s | 0.217s | helper phase archive |
| 2 | Seven npm directories | 5.736s | 0.315s | warm then fresh npm |
| 2 | TypeScript and 100 MiB file | 27.253s | 6.249s | warm then fresh corpora |
| 2 | Checker archive | 2.234s | 0.215s | warm then fresh archive |
| 3 | Seven npm directories | 7.500s | 0.214s | fresh then warm npm |
| 3 | TypeScript and 100 MiB file | 22.897s | 5.830s | fresh then warm corpora |
| 3 | Checker archive | 2.176s | 0.215s | fresh then warm archive |

Best warm: npm 0.214s, corpora 5.830s, archive 0.215s. The full TypeScript byte validation dominates warm time. Tools are already installed. The first archive sample needed an uncached validation compile and took 176.479s; later fresh directories reused those Go actions. Calling the 2.176s best archive sample completely cold provisioning would conceal compilation. Build flags for every sample: main 39638d9e278d38bb5aeae887f46d55a70e47aaad; nproc=5; cgroup cpu.max=400000 100000; Go 1.27.1 linux/amd64; clang 20.1.8 LLVM 87f0227cb60147a26a1eeb4fb06e3b505e9c7261; Node v24.19.0. Cache state and contemporaneous load before/after are recorded individually.

Initial branch ordinary setup took 28.745s; separate fresh main checkout setup 209.159s including submodules and Go exports. Initial flag setup waited 73s for the installer lock, installed corpora/npm, then exited with no space left on device before publishing the archive. Only the completed markdown-fresh proof disposable Go cache and duplicate LLVM were deleted, retaining its logs and checkout. Retry succeeded in 44.727s with partially warm actions/inputs, not a fully cold VM. Final setup validates all ten input stamps. Outer Go warming may conservatively rebuild after other jobs add cache actions; its timing lines remain in the logs. Ordinary setup does no gate preparation and clears the opt-in exports after a gate run.

## Mutants, checks and limitations

29 distinct mutants: omit kind, HEAD, dependency actions, Go environment/version, C compiler version, link/validation flags, helper, TypeScript commit/URL, ignore size/content; omit artifact names/bytes/modes/root mode/entries; omit npm lock/manifest/bootstrap/helper/Node version; omit npm installed names/bytes/modes/root mode; disable bootstrap integrity rejection; restore header-producing go list mode. Each is caught by its intended assertion. Omitting dependency actions also fails a real dirty-source rebuild. Header omission uses a cold unique package namespace so the clean-checkout assertion catches the generated file. Each mutant failure log is retained under mutants/.

Final commands: ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_gate_inputs.py: 5 tests OK; ADAMIC_MARKDOWN_SETUP_MODULE=/workspace/adamic/cloud/setup-gate-npm.py ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_markdown_setup.py: 6 tests OK; python3 cloud/gate-input-mutants.py: all 29 caught; ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_setup.py: 5 tests OK on serial rerun; go vet ./...: exit 0; bash -n, Python AST checks and git diff --check pass. An initial Git fixture assertion forgot to restore its pin and was corrected; its failure is retained. Concurrent ordinary integration conservatively rebuilt test binaries instead of meeting the skip assertion; its failure is retained and serial rerun passes. No product failure was called a flake or fixed.

The prototype go list emitted an untracked generated .h during setup. It was removed and the installer fixed. First requested outputs are retained. The real-tree-dependent gitignore row was repeated on clean main after removal and passed; other requested checks read src/compiler, stage1 or JSON rather than that header. Final main status is empty; no branch source was copied into the audit checkout. Uncovered: whole gate beyond the requested rows, ARM, fresh VM provisioning, and the absent-main test. No container snapshot boundary is observable. Human-readable log trailing whitespace is normalized; Go JSON event streams remain exact.
