Built three default-option native Adamic rule ports and the raw declaration-contract checker question.
Commits: claim d34f940b; implementation a6d8d0293467d1d99adb40e6cd8787dc7925c684; evidence is committed with this report.
Checks: wave byte/sanitizer/mutant test PASS 53.857s; bridge PASS 69.115s; filtered Node oracle PASS 20.208s.
Mutants: all three rule mutants exit 0 and fail only Go diagnostic bytes; seven ABI, nineteen frame, two request and one Node mutants are caught.
Not covered: full repository gate, every upstream option/fixture matrix, JSX/JavaScript corpora, suppression or edit application; pinned CLI discovery cannot lint .a paths.

## Selection and implementation

Branch codex/typeaware-wave-15 starts at origin/codex/tsgo-c-library,
0d540f413625f016f20fea39761c7b184f335de6, as requested for this wave rather
than the generic main base. Cohere stays pinned at
715ba94f3608a6500086b1076ce5cb7e51b836db; typescript-go stays at
8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md and the complete
README, VOLUME_REPORT and COVERAGE_REPORT before implementation.
Selection sums validation-volume/compiler-all.counts and repository-all.counts,
removes the 26 existing ports, then sorts descending combined counts with lexical
full-name ties. Positions 43, 44 and 45 are:

| Rule | Compiler | Repository |
| --- | ---: | ---: |
| @typescript-eslint/no-useless-default-assignment | 0 | 1 |
| @typescript-eslint/prefer-find | 1 | 0 |
| @typescript-eslint/require-array-sort-compare | 1 | 0 |

Fetched all origin heads, checked claim paths, named port files and full rule
names across stage1 on origin branches. Matches were count logs and inventory
metadata, with no implementation or prior claim. No rule was skipped. The claim
commit was pushed before any implementation file was written.

Each rule lives in its own .a file. wave_15_suite.a owns an isolated runner,
using the existing parser, owned diagnostic/fix/suggestion representations and
one checker program across all roots. Existing runners and rules are untouched.
The only shared implementation edit is the declaration-contract dispatch entry
in bridge/tsgo/checker/facts.go, two lines after gofmt. Its Go implementation and
Adamic decoder each have a new file named declaration_contract. No protected
emitter, lowering, native driver or oracle file was changed.

The new question returns the declared parameter type and contextual signatures'
declaration identity, parameter presence, symbol flags, rest markers and type
graph. It returns no lint verdict or repair. Adamic decides callback contract
reachability, destructuring defaults, string-array exemptions, filter/index
recognition, diagnostics and repairs. Direct checker tests compare the contextual
parameter identity and reject a wrong node kind and a question suffix.

The independent oracle is testdata/oracle_wave_15.go, built through an overlay
inside pinned cohere. It invokes all three unmodified production Run functions,
with their production program views and shared file cache. It imports no bridge
code. Both sides retain configured declaration roots, parse .a roots, sort full
canonical findings and preserve duplicate findings and the ordering of fixes and
suggestions. Repairs are proposed rather than applied.

## Agreement and sanitizer observations

TypeScript corpus: v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8,
using the recorded 77 src/compiler roots. Repository: the recorded 287-file
manifest, including 212 .a and 75 .ts files. New ports are outside that frozen
population. Both manifests, final source hashes, diagnostic hashes, compressed
canonical outputs and test logs are in validation-wave-15.

| Population | Findings | Full identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | --- |
| Ten generated controls | 19 | 3569 | PASS |
| Compiler, 77 roots | 2 | 5629 | PASS |
| Repository, 287 roots | 1 | 18704 | PASS |

Controls cover array unions, tuples, custom sort methods, comparator arguments,
string exemptions, literal and computed filter names, constant indexes, fractional
at indexes, nonnumeric at indexes, comment-preserving suggestions, ternary
filters, undefined defaults, optional parameter repairs, nested destructuring,
optional tuple members, plain arrays, callback contextual types, plain arrow
parameters, Unicode and CRLF spans. Native sanitizer stderr is empty on all three
populations. ASan instruments native/C memory, not the Go heap.

Initial comparisons caught incorrect type flag masks, then an initializer helper
that mistook the equals before a single-parameter arrow for a default. The rule
now uses the pinned checker flags and scans tokens inside the declaration to
recognize an initializer. A new plain-arrow control holds the correction. These
failed runs were not counted as agreement.

## Every mutant run and what caught it

| Rule mutant | Result | Catch |
| --- | --- | --- |
| default-assignment: deletion end +1 | exit 0, empty stderr | Go full finding/fix bytes, byte 2053 |
| prefer-find: replace filter name with filter in suggestion | exit 0, empty stderr | Go full finding/suggestion bytes, byte 620 |
| sort-compare: invert string-like exemption | exit 0, empty stderr | Go full finding bytes, byte 49 |

No compilation error or sanitizer failure counts as a rule mutant kill.
Querying a released program through declaration-contract panics 70 with
`invalid or released checker handle`. The ABI regression's retained-registry
mutant separately fails its stale-handle assertion.

The unchanged bridge gate ran these seven mutants: input byte length +1 and
output string length +1 each trigger ASan heap-buffer-overflow; retained released
handle fails the stale-handle assertion; querying the source-file position fails
independent Go bytes at byte 6; removed link opt-in fails the refusal expectation;
omitted output-buffer frees and a region result allocated on the heap each
trigger LeakSanitizer. Its normal C ABI probe covers 100 queries, buffers surviving
release, nonreused handles, and rejection of zero and stale handles. Its independent
normal sample compares 162 positions and 3261 bytes under ASan/UBSan/LSan.

The unchanged decoder gate ran nineteen malformed-wire mutants: empty,
length-negative, length-short, trailing, version, question, not-integer,
rounded-integer, noncanonical-integer, negative-natural, bad-boolean, zero-id,
negative-tuple, missing-root, missing-constraint, missing-element, duplicate-id,
missing-link-17 and missing-link-18. All panic 70 with their specified frame or
graph rejection. Wrong-kind and unknown-question request mutants exit 0 when
the guards are disabled and fail the required panic expectation. The filtered
Node oracle runs the one-byte stdout mutant, which its comparison catches.
Full details are preserved in bridge.log, facts.log and node.log.

## Native cost against Go

One optimized, quiet command pair per corpus after formatting, all checks and
builds excluded. These are single-run observations, not benchmark medians.
Both timed outputs still compare byte for byte.

| Corpus | Native process | Go process | Native / Go | Native load | Native run | Go load | Go run | Native queries |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 1.946020 s | 0.361830 s | 5.38x | 0.243986 s | 1.679748 s | 0.233536 s | 0.109548 s | 1303 |
| Repository | 0.266485 s | 0.123296 s | 2.16x | 0.061940 s | 0.200080 s | 0.063264 s | 0.048367 s | 527 |

Native aggregate adapter intervals: compiler 248097035 ns, repository
20894537 ns. They include checker work, serialization, C transfers, output
conversion and freeing, and are not isolated C crossing latency. Native remains
slower than production Go. No performance optimization is claimed.

## Commands and environment

`bash cloud/setup.sh > /workspace/wave-15-setup.log 2>&1` passed on the first
attempt. Timing lines: go ready 0s; clang ready 0s; node ready 0s; submodules
ready 0s; build cache warm 77s; done 77s. `nproc`: 5. Cgroup cpu.max:
400000 100000. Memory: 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0,
Linux amd64. Each build/test command sources /workspace/adamic-tools/env.sh.

Final targeted gate:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE15_ARTIFACTS=/workspace/wave-15-validation \
ADAMIC_WAVE15_REPOSITORY_MANIFEST=/workspace/wave-15-repository.manifest \
ADAMIC_WAVE15_COMPILER_MANIFEST=/workspace/wave-15-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave15AgreementAndMutants$' \
    -v -count=1 -timeout=30m > /workspace/wave-15-final.log 2>&1
# PASS 53.857s

go test -v -count=1 -timeout=15m ./bridge/tsgo/... \
    > /workspace/wave-15-bridge.log 2>&1
# PASS bridge 69.115s, checker 0.136s

go test -v -count=1 -timeout=10m ./internal/oracle \
    -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
    > /workspace/wave-15-node.log 2>&1
# PASS 20.208s; eight fixtures, also selected method_closures and generic_functions

go test ./stage1/cohere/typeaware \
    -run '^Test(SixPinnedFlags|PinnedTypeFlags|FactsDecoderGuards|InspectRequestRefusals)$' \
    -count=1 -v -timeout=10m > /workspace/wave-15-facts.log 2>&1
# PASS 45.807s

go test ./bridge/tsgo/checker -v -count=1 > /workspace/wave-15-checker-test.log 2>&1
# PASS 0.138s

go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave-15-vet.log 2>&1
# exit 0, empty output
```

Go formatting and git diff --check produce empty logs. Production cohere stdin
formatting is applied as patches and then verified stable for all five .a sources.
The pinned CLI rejects directly named .a paths even though sourceExtensions is in
tsconfig; lint.log preserves the failure. The stdin workaround gives the formatter
an imaginary .ts filepath while keeping both source and output files .a. It is
formatting evidence, not a claim of a successful full configured lint run.
No submodule upgrade or extra shared-file change is used to conceal that limit.

All test output went to log files without pipes. No full repository test gate,
full configured lint gate, complete upstream fixtures/options, JSX/JavaScript
population, suppression engine or applied edit engine is claimed. Supported
production default options on the two pinned populations and ten generated
controls are the measured scope. No pull request is opened.
