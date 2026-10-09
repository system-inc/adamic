# Split-fuzz evidence

Branch: `devtools/split-fuzz`. Base: `54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8`, the `origin/main` fetched at the start. The starting workspace was `73352e87`; the fetched main already contains the optional-field generator fix, so these are before/after measurements of the fetched main, not a timing comparison with its earlier 64-CPU whole-gate log.

All 15 original fuzz tests were measured individually before changes. The six reported slow parents are now 42 independently selectable leaves; the nine other tests remain intact, giving 51 fuzz gate units. No final unit exceeded 30 seconds on this instance. Explicit registration in `cmd/adamic-gate` makes these actual gate units rather than just subtest display names.

## Instrument and cold boundary

`measure.py` uses **Python `time.monotonic()`**, not GNU time (which is absent), to record process wall time including log collection. It starts a separate compiled test process for every anchored selector, serially, with `GOMAXPROCS=4`, `-test.parallel=4`, `-test.count=1`, `ADAMIC_GATE_UNCACHED=1`, and a new `XDG_CACHE_HOME` per selector. The Go action cache is explicitly held at the prepared cache so changing XDG does not rebuild Go accidentally. These are cold observation/native-runtime measurements after shared tool/dependency/Go build preparation, not empty-Go-action-cache checkout-to-result measurements. Every `Prepare` inside the selected parent's body, including native runtime compilation, is included in its wall time. Each child selector is also checked to run exactly its parent and itself.

Shared environment setup took **209.766 s** and is outside the unit measurements. The instance has cgroup `cpu.max = 400000 100000` (4 CPUs), although `nproc` reports 5. Linux 6.18.44, Go 1.27.1, clang 20.1.8, Node 24.19.0. Binary, source, and instrument SHA-256 values and pinned submodule commits are recorded in `environment.json`.

The seed batches own separate scratch directories. Undefined-number and override sources are generated once per parent invocation and read by both the selected batches and vocabulary assertion. Checkout preparation is once per runtime/reduction parent invocation. A separately selected child on another instance performs that setup again; the table includes that repeated cold cost. No child needs its sibling to run first.

## All tests

| Test | Preserved cases | Before wall (s) | After units | Largest cold unit wall (s) |
| --- | --- | ---: | ---: | ---: |
| TestOneSeedOneProgram | unchanged | 0.033 | 1 | 0.029 |
| TestGeneratedProgramsCheckAndLower | 60 seeds (1..60) | 6.178 | 12 | 0.613 |
| TestRegexProgramsPassTheChecker | 30 seeds (1..30) | 0.907 | 6 | 0.175 |
| TestOctoberFeaturesAppear | unchanged | 0.094 | 1 | 0.090 |
| TestShrinkKeepsOnlyWhatFails | unchanged | 0.013 | 1 | 0.014 |
| TestSharedSliceCutsShare | unchanged | 0.019 | 1 | 0.020 |
| TestOwnershipShapes | unchanged | 0.853 | 1 | 0.839 |
| TestJudgeReadsThePanicLine | unchanged | 0.008 | 1 | 0.009 |
| TestOverridesShapesAndLower | 50 seeds (0..49), vocabulary | 1.917 | 11 | 0.226 |
| TestSourceTreeCutsAtItems | unchanged | 0.008 | 1 | 0.008 |
| TestReduceKeepsTheSignature | 1 fixture, 21 candidates, reobservation | 10.143 | 2 | 8.580 |
| TestExecuteCPUHelper | unchanged | 0.009 | 1 | 0.012 |
| TestExecuteCPULimit | unchanged | 4.692 | 1 | 4.668 |
| TestFuzzerSharesRuntimeLibrary | 1 library comparison, 1 program | 8.091 | 2 | 9.007 |
| TestUndefinedNumbersShapes | 40 seeds (1..40), vocabulary | 3.995 | 9 | 0.565 |

The after column is the maximum of separate cold child invocations, including parent preparation, rather than the child's Go-reported duration alone. `after.json` records every leaf, including each seed range and the two vocabulary leaves. `before.json` records all original parents. The whole-gate historical values are not comparable to these isolated runs, and these results do not claim a whole-gate speedup.

## Preservation and planted failures

`prove.py` uses Go source overlays, leaving the checked-out sources untouched. It logs the seed and SHA-256 immediately before the original and new tests write the program consumed by the checker. The disjoint unions are exactly **60 + 30 + 40 + 50 = 180** checked programs: every original seed occurs once and every generated-source hash matches. The regression test `TestFuzzChildrenCoverOriginalSeeds` separately checks the gate's enumeration and the literal range bounds for missing, duplicate, and mislabeled seeds.

The undefined-number vocabulary retains all **24 shape definitions** (23 permitted, one opt-in exclusion), all **7 source kinds**, the omitted-interface opt-in search, and the disabled-feature check. Overrides retain all **29 markers** and the disabled-feature check. The input hashes make the vocabulary's source corpus identical. The vocabulary helper lives outside the undefined-number test so the gate's AST reader does not mistake its source-kind list for subtest names.

The reducer's before/after operation multisets match: one preparation, one original observation, one derivation, **21 candidate observations with identical source hashes**, one reduction result, and one final observation with the original minimum's hash. The runtime test retains one preparation, one runtime-library comparison and one three-backend program. See `inputs.json` and `checkout-inputs.json`.

Ten mutants were caught only by the intended leaf; all its siblings passed. The seed mutants append a real invalid TypeScript declaration (`const splitFailure: number = 'planted';`) to one program, producing TS2322. The vocabulary mutants require a missing source kind or marker. Runtime mutants change the real program output or the comparison library's sanitization flags. Reducer assertion mutants change the expected signature or expected minimum; the reduction algorithm and its input remain intact.

| Mutant | Only failing child | Passing siblings |
| --- | --- | ---: |
| TestGeneratedProgramsCheckAndLower-seed-13 | TestGeneratedProgramsCheckAndLower/seeds-011-015 | 11 |
| TestRegexProgramsPassTheChecker-seed-18 | TestRegexProgramsPassTheChecker/seeds-016-020 | 5 |
| TestUndefinedNumbersShapes-seed-28 | TestUndefinedNumbersShapes/seeds-026-030 | 8 |
| TestOverridesShapesAndLower-seed-43 | TestOverridesShapesAndLower/seeds-040-044 | 10 |
| TestUndefinedNumbersShapes-vocabulary | TestUndefinedNumbersShapes/vocabulary | 8 |
| TestOverridesShapesAndLower-vocabulary | TestOverridesShapesAndLower/vocabulary | 10 |
| runtime-program | TestFuzzerSharesRuntimeLibrary/program | 1 |
| runtime-library | TestFuzzerSharesRuntimeLibrary/library | 1 |
| reduce-signature | TestReduceKeepsTheSignature/signature | 1 |
| reduce-minimum | TestReduceKeepsTheSignature/reduction | 1 |

## Validation and reproduction

Passed: `gofmt -l cmd internal` (empty), `git diff --check`, `go vet ./...`, `go test -count=1 -timeout=30m ./cmd/adamic-gate`, and the two complete touched packages with a fresh native cache, uncached observations, and parallelism 4. The combined package run reports `internal/fuzz` 15.167 s; its largest parent event is 9.51 s. The complete repository test suite and an actual run on 100 instances were not performed. A 100-shard plan is checked after committing the clean checkout; its full evidence is kept outside the repository.

```sh
source /workspace/adamic-tools/env.sh
python3 review/split-fuzz/measure.py before /tmp/adamic-split-reproduce
python3 review/split-fuzz/prove.py /tmp/adamic-split-reproduce
python3 review/split-fuzz/measure.py after /tmp/adamic-split-reproduce
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test -count=1 -parallel=4 -timeout=30m ./internal/fuzz ./cmd/adamic-gate
go vet ./...
go run ./cmd/adamic-gate plan -count 100 > /tmp/adamic-split-plan.json
```

The baseline instrument overlays the pinned base's original test sources. Run it on this branch, whose generator/compiler sources match that base. Complete raw process and mutant logs are under `/tmp/adamic-split-fuzz` on the measured instance; they are not committed, following the repository's rule for test logs. Structured timings, inventories and mutant outcomes are committed alongside the instruments.
