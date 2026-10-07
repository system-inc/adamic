# Refusal coverage measurements

Writer commit `1ac8ea246a0ee89116c7ac72b02b0e914e845862`, based on area/developer-tools `3f904d5364d39c9a1d03845fe2981537d5227561`. Target compiler: current origin/main at fetch, `39638d9e278d38bb5aeae887f46d55a70e47aaad`. No compiler fixes were made. Separate worktrees began at that target; only the probe code and one catalog patch were added. The control has no compiler diff.

## Main, seed 1

| Result | Programs |
|---|---:|
| refused-as-expected | 1678 |
| accepted | 160 |
| wrong-refusal | 0 |
| not-yet | 162 |
| load-error | 0 |
| invalid-neighbor | 0 |

All 322 findings, each with its complete program, accepted neighbor, index, seed, surrounding and expected/actual diagnostic, are listed in [main-2000.jsonl](results/main-2000.jsonl). This is the list of every finding, not just representative programs. Findings are not suppressed or used to fix the compiler.

| Construct | Kind | Findings |
|---|---|---:|
| any | not-yet | 41 |
| eval | not-yet | 40 |
| expando | not-yet | 41 |
| function-type | accepted | 40 |
| merging | accepted | 40 |
| new-function | not-yet | 40 |
| optional-widening | accepted | 40 |
| record | accepted | 40 |

Observed: unused `Function` parameter declarations, unused `Record` annotations, optional widening and class/interface merging compile without the required refusal. Explicit `any`, function expandos, eval and new Function return NotYet instead of Refused. These establish missing refusal/diagnostic coverage. The accepted cases include unused declarations and are not independent demonstrations of runtime type corruption. No execution result is inferred from compilation alone.

## Regression detection

Patch 11 applied unchanged. Patch 12 failed git apply --check because newer main also uses fmt/scanner for checking pragmas. Adaptation: delete only the 11-line CommentDirectives block; retain the pragma block and its imports. The resulting diffs are [patch-11.diff](results/patch-11.diff) and [patch-12.diff](results/patch-12.diff).

Each pair uses the same seed and selected construct sequence. Controls check 200 programs; mutants stop at the first finding. Three interleaved before/after pairs were run per patch with prebuilt binaries. The first finding is program 1 in every mutant run. The controls all exit 0, 200 refusals as expected. The full main corpus is not clean, as reported above; the affected regression seeds have clean controls.

| Loop | Before | After | Instrument (exact command) |
|---|---|---|---|
| patch 11, best of 3 | clean, 200 programs, 10.033642 s whole run | accepted at program 1, 0.056407 s to finding | `ADAMIC_GATE_UNCACHED=1 /tmp/refusal-{main,11}-bin -root /tmp/refusal-{main,11} -seed 1 -count 200 -only definite-local,definite-field`; append `-stop` only for mutant |
| patch 12, best of 3 | clean, 200 programs, 9.653030 s whole run | accepted at program 1, 0.049749 s to finding | `ADAMIC_GATE_UNCACHED=1 /tmp/refusal-{main,12}-bin -root /tmp/refusal-{main,12} -seed 1 -count 200 -only ts-ignore,ts-expect-error`; append `-stop` only for mutant |
| main corpus | no earlier writer baseline | 119.713405 s, 2,000 programs | `ADAMIC_GATE_UNCACHED=1 /tmp/refusal-main-bin -root /tmp/refusal-main -seed 1 -count 2000 -out /tmp/refusal-final-findings` |

The command pair notation above expands to the literal commands stored per run in [measurements.json](results/measurements.json). No compiler speedup is claimed: before is a complete clean control and after is time to the first failure. Writer detection clocks start after argument handling/catalog auditing; whole-process wall durations are also stored. The binary is built before the clock, so these are warm-build detection measurements, not time from building a checkout. Each first finding follows checking its accepted neighbor.

## Build flags

`target_commit=39638d9e278d38bb5aeae887f46d55a70e47aaad; writer_commit=1ac8ea246a0ee89116c7ac72b02b0e914e845862; nproc=5; cpu.max=400000 100000; go=go version go1.27.1 linux/amd64; clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; build=go build -trimpath -buildvcs=false; cached=no probe cache; ADAMIC_GATE_UNCACHED=1`

Every run carries that build-flags object and its own load-before/load-after in measurements.json. Per-run timing logs are separate from deterministic JSON. No probe cache exists. Default and ADAMIC_GATE_UNCACHED=1 over seed 1, count 200 have byte-identical stdout, including all diagnostics and findings (mode-default.jsonl versus mode-uncached.jsonl); both exit 1 because of the real findings.

Initial setup printed Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 118s, done 118s. That older setup invocation did not capture load averages. A second current-branch setup captured the required complete metadata in [setup.log](results/setup.log); its timing lines are reproduced there, not treated as an interleaved speed comparison.

## Verification and mutants

Tests passed: `go test -count=1 ./internal/refusalprobe ./cmd/adamic-refusals`; every executable entry occurs in the first 200 programs, and every one of those 200 repaired surroundings compiles. Source-map, direct-refusal and helper auditing fail for unknown entries. `go vet ./...`, `gofmt -l cmd internal`, and `git diff --check` passed.

Focused lowering: `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/lower -run 'TestDefiniteAssignment|TestSuppressionDirective|TestCheckPragma|TestWhatZeroOneRefuses|TestGeneratorsAreRefused'` passed.

Filtered oracle: `ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^load$/^testdata$/^0.1$/^compile$/^01_hello[.]ts$'` passed, actually executing the fixture (native misses 3, Node misses 2, no cache hits). Every test run wrote to a log, never a pipe. The full gate was not run.

| Mutant | Check that failed |
|---|---|
| accept any Refused diagnostic | TestWrongDiagnosticIsAFinding/wrong and TestStrictDiagnosticWithRealLowering |
| skip neighbor compilation | TestNeighborFailureIsNotACompilerHole |
| omit the first executable catalog entry | TestCatalogCoverage, missing any |
| treat every diagnostic as catalog-covered | TestCatalogAuditCanFail and TestDirectRefusalAuditCanFail |
| catalog patch 11 | writer: definite-local accepted, program 1 |
| adapted catalog patch 12 | writer: ts-ignore accepted, program 1 |

The four test mutants failed their intended assertions, not compilation or the toolchain. Logs are in results/*-mutant.log. An initial run of the extra three mutants used the system non-Go executable and failed at invocation; those invalid observations were discarded and rerun with the sourced toolchain. No result caches were added, so no cache-key mutants apply.

## Limits

This is partial coverage, explicitly represented by catalog Boundary entries and in every command summary. The first-200 guarantee applies to executable entries, not boundaries. In particular yield requires a generator (a second refused construct); with and enums cannot reach lower.Refused through the strict loader. The original 0.1 refusal table includes features opened in later compiler revisions. Current main cannot simultaneously accept their neighbors and refuse every old syntax as requested.

Not covered: exhaustive delegated-helper refusal cases, inheritance override scenes, overloads, remaining library omissions, multi-module refusals/cycles and polymorphic recursion. Those omissions are limitations of this writer, not claims of impossible syntax. Runtime differential tests for the accepted declarations are also not covered. The catalog audit is exhaustive for both maps and direct Refused expressions in refusals.go, but not for the bodies of delegated helper functions.
