Built: rebased the existing wave-18 branch onto current main and re-greened its oracle suites; no new claims.
Commits: previous remote 72f84449; tested rebased code 72cb3b4f; base e8ba3d5d.
Checks: all five wave-18 suites passed in 420.094s; checker 0.144s; uncached Node oracle 2.144s; vet exit 0.
Mutants: fifteen full-suite byte mutants, four stale-handle mutants and six partial reporter/refusal mutants caught.
Not covered: native React HIR/SSA analysis, base-parser JSX integration, full root gate or new rules.

Only codex/typeaware-wave-18 was pushed by this unit. It was not on main at the
start. All 22 commits rebased without conflicts. The first validation on e011f8f6
passed all five suites in 407.713s, but main advanced during the run. Its new
call-target and devirtualization changes affect native generation and memory
analysis, so the branch was rebased again and the checks repeated. The final
fetch confirmed main was still e8ba3d5d, an ancestor of the tested branch.

The rebase did not change owned rules or bridge-question sources. No shared
parser, harness, registration generator or protected compiler files were manually
edited. The latest evidence commit changes only owned reports, claim status and
validation artifacts. Older report SHAs identify pre-rebase commits; range-diff.txt
maps them to their rebased equivalents.

## Commands and outputs on the final base

Sourced /workspace/adamic-tools/env.sh; TMPDIR=/tmp/adamic-gate. Setup printed
Go ready 0s, clang ready 0s, Node ready 0s, submodules 1s, build cache warm 118s,
done 119s; nproc=5, cpu.max=400000 100000. Toolchain: Go 1.27.1, clang 20.1.8,
Node 24.19.0. CoHere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.

    ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript
    ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest
    ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest
    ADAMIC_WAVE18_JSX_SLICE=/workspace/wave-18-jsx-slice
    ADAMIC_WAVE18_ARTIFACTS=/workspace/wave18-landing-current/original
    ADAMIC_WAVE18_TITLE_ARTIFACTS=/workspace/wave18-landing-current/title
    ADAMIC_WAVE18_CORE_ARTIFACTS=/workspace/wave18-landing-current/core
    ADAMIC_WAVE18_CONSTRUCTOR_ARTIFACTS=/workspace/wave18-landing-current/constructor
    ADAMIC_WAVE18_PREFERENCE_ARTIFACTS=/workspace/wave18-landing-current/preference
    go test ./stage1/cohere/typeaware -run '^TestWave18' -count=1 -timeout=30m -v

These variables were exported. Tests wrote to /tmp/wave18-landing-current-rule-tests.log.
Every suite passed, including title with its real native JSX slice, with no skip:

* TestWave18ConstructorAgreementAndMutants: 103.03s.
* TestWave18CoreAgreementAndMutants: 80.32s.
* TestWave18PreferenceAgreementAndMutants: 95.40s.
* TestWave18AgreementAndMutants: 88.55s.
* TestWave18TitleWithJsxSlice: 52.79s.

Normal and ASan/UBSan streams agree on full findings, fixes and ordered suggestions.
Both frozen corpora (77 compiler and 287 repository sources) agree in each suite.
Options, no-library and ambient cases are included where the existing suite tests
them. Each source bridge also passes its released-handle checks. Corpus source
hashes are recorded; absolute artifact headers account for differing byte lengths
from older runs.

    go test ./bridge/tsgo/checker -count=1 -timeout=10m
    go vet ./...

Checker output: ok, 0.144s. Vet exits 0 with no diagnostics. The Node command was:

    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)\.a$' -count=1 -timeout=15m -v

It passed in 2.144s, with native hits=0/misses=37 and node hits=0/misses=25.
The suffix regex additionally selects regexp_cycle_closures.a. It includes the
new call-target/devirtualization fixtures as well as the one-byte oracle mutant.
All output went directly to logs. The earlier cached run is retained separately.

    python3 stage1/cohere/typeaware/wave_18_react_partial/validate_partial.py --scratch /workspace/wave18-landing-current/react-partial

Reporting only: four production Go findings match 2423 bytes, normal and ASan/UBSan.
The three diagnostic-ID mutations compile, exit 0 with empty stderr and fail Go byte
comparison. The three removed-refusal mutations change expected exit 70 to exit 0.
These are reporting/refusal mutants, not completed source-analysis mutants.

## Every full-suite mutant rerun

The fifteen byte mutants compile, exit 0 with empty stderr and are rejected only
by the independent production Go finding comparison. Four registry mutants retain
a released program, causing exit 0 where the real bridge panics with exit 70.
The complete test evidence is:

* wave_18_constructor_test.go:126: new-func mutant: exit 0, empty stderr, Go byte oracle caught byte 1144
* wave_18_constructor_test.go:126: native-nonconstructor mutant: exit 0, empty stderr, Go byte oracle caught byte 11761
* wave_18_constructor_test.go:126: new-wrappers mutant: exit 0, empty stderr, Go byte oracle caught byte 13702
* wave_18_constructor_test.go:169: released registry mutant caught: expected stale handle panic, got <nil>
* wave_18_core_test.go:120: class-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 58
* wave_18_core_test.go:120: const-assign mutant: exit 0, empty stderr, Go byte oracle caught byte 571
* wave_18_core_test.go:120: constant-binary mutant: exit 0, empty stderr, Go byte oracle caught byte 42582
* wave_18_core_test.go:163: released registry mutant caught: expected stale handle panic, got <nil>
* wave_18_preference_test.go:144: promise mutant: exit 0, empty stderr, Go byte oracle caught byte 1923
* wave_18_preference_test.go:144: regex mutant: exit 0, empty stderr, Go byte oracle caught byte 34564
* wave_18_preference_test.go:144: rest mutant: exit 0, empty stderr, Go byte oracle caught byte 23201
* wave_18_preference_test.go:154: regex grammar mutant: exit 0, empty stderr, byte oracle caught byte 34476
* wave_18_preference_test.go:197: released registry mutant caught: expected stale handle panic, got <nil>
* wave_18_test.go:123: await mutant: exit 0, empty stderr, Go byte oracle caught byte 62
* wave_18_test.go:123: class mutant: exit 0, empty stderr, Go byte oracle caught byte 4539
* wave_18_test.go:139: iteration-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 1752
* wave_18_test.go:139: base-facts mutant: exit 0, empty stderr, Go byte oracle caught byte 7962
* wave_18_test.go:195: released registry mutant caught: expected stale handle panic, got <nil>
* wave_18_title_test.go:105: title mutant exit 0 empty stderr: byte oracle caught byte 61

## Native time against Go

Whole-process median seconds from three interleaved rounds, as observed during
validation. These are corpus load/parse/lint count timings, not a general speedup
claim. Native remains slower than Go. The React reporters are excluded because
formatting supplied spans is not comparable to Go source analysis.

| Suite | Corpus | Native seconds | Go seconds |
| --- | --- | ---: | ---: |
| wave_18_constructor_test.go | repository | 0.243514 | 0.131333 |
| wave_18_constructor_test.go | compiler | 1.601408 | 0.315175 |
| wave_18_core_test.go | repository | 0.313747 | 0.146680 |
| wave_18_core_test.go | compiler | 2.149176 | 0.415157 |
| wave_18_preference_test.go | repository | 0.246773 | 0.148138 |
| wave_18_preference_test.go | compiler | 2.034954 | 0.390341 |
| wave_18_test.go | repository | 0.232280 | 0.128450 |
| wave_18_test.go | compiler | 1.461898 | 0.305964 |
| wave_18_title_test.go | repository | 0.232355 | 0.134330 |
| wave_18_title_test.go | compiler | 1.394647 | 0.299429 |

## Remaining boundary

All three React source-analysis rules remain unfinished: set-state-in-effect,
set-state-in-render and static-components need native React HIR/SSA, captures,
compilation-unit selection and dominance; effect also needs memoization erasure
and inlining. The shared .a loader/profile/suggestion work does not supply these
analyses. Existing reporters refuse analysis explicitly and are not registered as
completed rules. The native base-parser JSX probe still exits 70. The Next title
visitor is validated against its isolated published native parser dependency;
its base-parser integration is still outstanding. Details remain in
WAVE_18_REACT_BLOCKER_REPORT.md, wave_18_react_partial/REPORT.md and
WAVE_18_TITLE_REPORT.md. No additional rules were claimed.

The full root gate was not run. The touched checker package, every owned completed
rule oracle, partial reporting/refusal checks and the filtered uncached Node oracle
were rerun. Complete stdout/stderr transcripts are gzip-compressed without truncation
under validation-wave-18-landing; source manifests, timing medians and rebase mapping
are included. No PR was opened.
