Built: landing rebase of all wave-18 work onto current main f8013f0b, with six oracle suites re-green.
Commits: rebased code 7015f8e9; prior pushed tip ac0a8b903; evidence commit follows.
Checks: five suites pass in the 797.277s combined attempt; corrected title suite PASS 51.778s; bridge checker, vet and filtered uncached Node pass.
Mutants: 19 native rule/fact/core mutations caught only by Go bytes, four stale-registry mutations caught, and six React reporter/refusal mutations caught.
Not covered: native React source-to-HIR/SSA/gates, full root gate, deterministic Go ambiguous-creator parity, or migration to an unspecified shared Diagnostic SHA.

The branch rebased cleanly onto f8013f0b. A final fetch confirms that main
remains unchanged and is an ancestor. Only codex/typeaware-wave-18 is a push
target. An exact lease on its prior ac0a8b903 tip protects the requested rebase
update. No new claims were made: three React source analyses remain unfinished.
Their rule.json manifests retain numeric kinds [307]. The origin audit examined
508 refs and found no native source-to-HIR adapter for these validators.

## Commands and results

    git rebase origin/main
    bash cloud/setup.sh > /tmp/wave18-f801-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v > /tmp/wave18-f801-oracle-final.log 2>&1
    go test ./stage1/cohere/typeaware -run '^TestWave18TitleWithJsxSlice$' -count=1 -timeout=30m -v > /tmp/wave18-f801-title-final.log 2>&1
    go test ./bridge/tsgo/checker -count=1 -timeout=10m
    go vet ./...
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)\.a$' -count=1 -timeout=15m -v
    python3 stage1/cohere/typeaware/wave_18_react_partial/validate_partial.py --scratch /workspace/wave18-f801/partial

The first lint attempt omitted ADAMIC_TYPESCRIPT_SOURCE and stopped at the
compiler corpus's invalid-config panic; it was terminated and preserved.
The corrected combined attempt passed five suites but failed title compilation
because ADAMIC_WAVE18_JSX_SLICE included stage1/typescript twice. Correcting
that environment path and rerunning only the title suite passed. Neither
configuration failure is counted as a green full run, and neither required
source changes. The initial checker path was incorrect, and the initial Node
selectors did not select the intended fixtures; corrected runs are authoritative.

Final environment: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript,
ADAMIC_WAVE18_JSX_SLICE=/workspace/wave-18-jsx-slice, compiler config
/workspace/wave-18-typescript/src/compiler/tsconfig.json. Manifest variables
point to /tmp/wave-18-{compiler,repository}.manifest. All six artifact variables
point to separate directories under /workspace/wave18-f801-final. The React
prepared-input config variable points to the same compiler config. Exact
transcripts and source hashes are preserved in validation-wave-18-f801.

| Suite | Seconds |
| --- | ---: |
| TestWave18ConstructorAgreementAndMutants | 75.09 |
| TestWave18CoreAgreementAndMutants | 82.97 |
| TestWave18PreferenceAgreementAndMutants | 96.85 |
| TestWave18ReactCores | 452.53 |
| TestWave18AgreementAndMutants | 86.17 |
| TestWave18TitleWithJsxSlice | 51.77 |

Both frozen corpora, 77 compiler sources and 287 repository sources, match full
findings/fixes/suggestions normally and under sanitizers. Title uses the isolated
native JSX parser; React uses Go-prepared HIR and does not prove native source
lowering. React retains 91 valid controls and 48 findings, with nine explicit
parse exclusions. Reporter validation matches four findings in 2355 bytes,
normally and under ASan/UBSan. The three ID mutants fail only Go-byte comparison;
the three removed-refusal mutants exit 0 instead of required 70. Base JSX
source analysis still refuses. No new bridge query was introduced by this rebase.

Bridge checker PASS 0.327s includes ownership/released-handle contracts. Vet
exits 0. Filtered Node PASS 8.412s, native hits 0/misses 37 and Node hits 0/misses
25, includes the one-byte oracle mutation. Setup prints Go 0s, clang 0s,
Node 0s, submodules 0s, cache warm 106s, done 106s; nproc 5, CPU quota 4.

## Every native mutation

All 19 byte-only mutations compile, exit 0 and have empty stderr. The four
released-registry mutations make a stale handle live and violate the stale
handle panic contract. Exact observations follow, including differing bytes:

- wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1124
- wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11531
- wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13402
- wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 53
- wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 566
- wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 42027
- wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1773
- wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 33554
- wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 22641
- wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 33466
- wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_react_cores_test.go:176: render-unconditional: exit 0, empty stderr; independent complete-byte comparison catches byte 448
- wave_18_react_cores_test.go:176: effect-setter: exit 0, empty stderr; independent complete-byte comparison catches byte 1549
- wave_18_react_cores_test.go:176: static-creator: exit 0, empty stderr; independent complete-byte comparison catches byte 56
- wave_18_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3115
- wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 57
- wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4474
- wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1722
- wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7877
- wave_18_test.go:195: released registry mutant caught: expected stale handle panic, got <nil>
- wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 66

## Native time against Go

Three alternating-order process measurements per suite/corpus yield the medians
below, in seconds. Native is slower on these source pipelines; no speedup claim.

| Suite/corpus/backend | Median seconds |
| --- | ---: |
| wave_18_constructor_test.go/compiler/go | 0.312284 |
| wave_18_constructor_test.go/compiler/native | 1.593252 |
| wave_18_constructor_test.go/repository/go | 0.126742 |
| wave_18_constructor_test.go/repository/native | 0.238747 |
| wave_18_core_test.go/compiler/go | 0.417852 |
| wave_18_core_test.go/compiler/native | 2.268312 |
| wave_18_core_test.go/repository/go | 0.153483 |
| wave_18_core_test.go/repository/native | 0.323542 |
| wave_18_preference_test.go/compiler/go | 0.410907 |
| wave_18_preference_test.go/compiler/native | 2.124120 |
| wave_18_preference_test.go/repository/go | 0.134899 |
| wave_18_preference_test.go/repository/native | 0.253577 |
| wave_18_test.go/compiler/go | 0.289167 |
| wave_18_test.go/compiler/native | 1.388694 |
| wave_18_test.go/repository/go | 0.131013 |
| wave_18_test.go/repository/native | 0.238047 |
| wave_18_title_test.go/compiler/go | 0.291643 |
| wave_18_title_test.go/compiler/native | 1.411874 |
| wave_18_title_test.go/repository/go | 0.123544 |
| wave_18_title_test.go/repository/native | 0.228550 |

React's prepared-input native total is 0.017811s across controls; Go source
analysis totals 0.294959s and Go fixture preparation 0.292960s. These perform
different work and are not comparable end-to-end timings. Existing numeric
listener declarations and manifests remain consistent, but the native source
adapter and supplied-node dispatch are not implemented. Older wave-18 rules
still use the shared parser's string-kind API. No shared parser, harness,
registration generator or protected compiler file was changed.
