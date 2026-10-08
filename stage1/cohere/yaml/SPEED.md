# YAML native speed work

## Shared schema checkpoint

The composer compiled both fixed schema tables on every formatting call.
It now shares their completed pattern trees for the module's lifetime. Matching
only reads those trees; document state and diagnostics remain per composer.
An initial release comparison measured 3.555806s before and 1.267965s after
(2,819.61 to 7,907.16 texts/s), with all 992,282 saved Go answer bytes unchanged.
These single observations are not the final five-round speed comparison.

The composer/schema/formatter/driver/mutant suite passed in 156.388s, including
sanitized native, source Node, emitted JavaScript and independent pinned libraries.
All 36 repository files and 10,026 formatter cases still match Go; the same 42
independently proved published-Prettier differences remain. Nine existing mutants
(composer, schema and printer) compile and finish with empty stderr; byte
comparisons catch each. Cohere's 276-rule check reported 42 files and 100%
Adamic-ready, and formatting passed.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=20m ./stage1/cohere/yaml -run 'TestCompose|TestSchema|TestFormatter|TestFileDriver|TestBundledParser' > /tmp/stage1-yaml-speed/shared-schema-suite.log 2>&1
```

Logs: [suite](audit/speed-shared-schema-suite.log),
[lint](audit/speed-shared-schema-lint.log),
[format](audit/speed-shared-schema-format.log).

## Runtime/compiler cost proving program

[gaps/stringUnitScan.ts](gaps/stringUnitScan.ts) computes the same UTF-16 checksum
using either one-unit slices or direct numeric unit reads. It consumes a file,
so the scanned strings are built at runtime. On `('abc中😀' repeated 4000)` and
100 rounds, both native variants and source Node print the same checksum.
The counted native build reports:

```text
slice:   allocations 2400007 frees 2400007 retains 105 releases 2400110 peak 7 regions 0
numeric: allocations       7 frees       7 retains 105 releases     110 peak 7 regions 0
```

Observed: 2.4 million unit slices make 2.4 million allocations and releases;
the numeric reads avoid them. Runtime `string.c` clamps and maps both UTF-16
slice boundaries; `string_share.c` copies slices shorter than 64 bytes, except
whole-string slices. This explains this probe's allocation cost. It is a
performance cost, not a demonstrated correctness bug. Compiler/runtime files
are outside this unit's scope and remain unchanged.

```sh
go run ./cmd/adamic build stage1/cohere/yaml/gaps/stringUnitScan.ts -o /tmp/stage1-yaml-speed/unit-scan --count
/tmp/stage1-yaml-speed/unit-scan /tmp/stage1-yaml-speed/unit-scan.txt slice 100
/tmp/stage1-yaml-speed/unit-scan /tmp/stage1-yaml-speed/unit-scan.txt numeric 100
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/yaml/gaps/stringUnitScan.ts /tmp/stage1-yaml-speed/unit-scan.txt slice 100
```

The existing [shared slice append correctness gap](gaps/sharedSliceAppend.ts)
remains separately held by its existing test. This optimization does not use
string append or alter that runtime behavior.

## Linux sampling profile

GNU gprofng 2.44 samples the unmodified release binary at 100ms intervals on
Linux 6.18.44. The 10,026-case corpus is repeated ten times, giving 100,260
formatting calls and 360 weighted ticks. After removing only gprofng's experiment
banner, all 9,922,820 output bytes equal ten copies of the saved Go answers;
target stderr is empty. No compiler/runtime source was edited for profiling.

| Baseline exclusive cost | Sample-weighted seconds | Percent |
| --- | ---: | ---: |
| `adamic_release` | 5.5 | 15.28% |
| `adamic_string_slice` | 4.3 | 11.94% |
| `adamic_allocate` | 3.9 | 10.83% |

`SchemaPattern_atom` appears on 50.83% of sampled stacks (inclusive, overlapping
with the exclusive costs above). This identified repeated schema construction.
The final fast-path profile has 10.3 sample-weighted seconds against the baseline's
36.0; it is not a substitute for the final unprofiled five-round benchmark.

Sampling is coarse. gprofng emits a timer-period-changed warning (`100000 -> 0`)
near collector shutdown; its weighted 36.0 seconds agree with the 36.1066-second
collection duration. Earlier 1ms requests were under-sampled (365 weighted ticks
in 36.56 seconds) and are not used for CPU-second estimates. Both runs ranked
release, slicing and allocation first; percentages vary with sampling noise.
The final 100ms profile has only 103 weighted ticks, so its finer ranking is
provisional. Headers are preserved rather than hiding collector warnings.

`perf` was not installed. Attempted apt installation failed because this worker
is uid 1000 and cannot acquire `/var/lib/dpkg/lock-frontend` or write
`/var/lib/apt/lists/partial`; sudo is absent. Installed gprofng supplied the Linux
sampling profile without installing packages.

```sh
gprofng collect app -p 100 -o /tmp/stage1-yaml-speed/baseline-100ms.er /tmp/stage1-yaml-speed/baseline --cases /tmp/stage1-yaml-speed/profile-cases.txt > /tmp/stage1-yaml-speed/baseline-100ms.stdout 2> /tmp/stage1-yaml-speed/baseline-100ms.stderr
gprofng display text -functions /tmp/stage1-yaml-speed/baseline-100ms.er > /tmp/stage1-yaml-speed/baseline-100ms.txt
gprofng display text -header /tmp/stage1-yaml-speed/baseline-100ms.er > /tmp/stage1-yaml-speed/baseline-100ms-header.txt
```

Logs: [baseline functions](audit/speed-baseline-profile.log),
[baseline header](audit/speed-baseline-profile-header.log),
[fast-path functions](audit/speed-fast-path-profile.log),
[fast-path header](audit/speed-fast-path-profile-header.log).

## Matcher and printer fast paths

Schema tests compile possible first-unit masks, including successors of nullable
concatenation prefixes. Impossible inputs return false; possible inputs and empty
strings still use the same anchored matcher. The recursive matcher reads its
stable arena directly instead of obtaining an owning node from a getter.

The pinned emoji RE2 graph is indexed by its first consumed mapped UTF-16 unit.
Epsilon paths preserve depth-first branch priority, and matching resumes at the
original consuming instruction. Only units absent from that exact table skip
matching. The printer uses a named child-print method rather than allocating a
closure per node. ASCII escape/whitespace tests use numeric reads.

All 36 files, 10,026 format cases, 50 direct driver cases and 9,388 standalone
schema cases remain unchanged against Go and their independent libraries.
`TestWidthsMatchGo` adds every Unicode scalar value and 256 emoji/context
sequences: 1,112,320 cases, 2,224,684 Go-identical answer bytes on sanitized native,
source Node and emitted JavaScript. It does not claim every possible emoji string
or malformed UTF-8 is covered.

The full YAML suite passed in 316.235s. The separately added cost-probe suite
passed in 2.585s on native with ASan/UBSan/LSan, source Node and emitted JavaScript.
The uncached filtered compiler oracle passed in 1.777s, with 27 native and 18 Node
misses, zero hits. Vet/gofmt logs are empty. Cohere reports 276 rules, 43 files,
100% Adamic-ready, and formatting passes. The complete repository gate was not run.

The three new qualifying mutants compile and finish with empty stderr on both
native and Node; only byte comparison catches them:

| Mutant | Held comparison | First differing byte |
| --- | --- | ---: |
| Nullable schema prefix excludes following first units | Schema vs Go | 6953 |
| Batch backslash numeric test recognizes CR instead | Formatter vs Go | 79 |
| Emoji first-unit range omits its endpoint | Formatter vs Go | 865030 |

All 24 earlier port mutants still pass their detection tests, for 27 qualifying
mutants total. An attempted initial-branch-order reversal survived the formatter
corpus. Further inspection proved it equivalent for this pinned graph: 11 initial
consuming instructions cover 185 mapped first units, with exactly one candidate
per unit and no competing units. It is not credited as a caught mutant.

Counted before/after builds reproduce Go's bytes:

```text
before: allocations 50991977 frees 50991977 retains 79230069 releases 103985985 peak 66662 regions 0
after:  allocations  6915652 frees  6915652 retains 26942422 releases  26333298 peak 66894 regions 0
```

Allocation/frees fell 86.4%. `peak` counts live allocations, not bytes or RSS;
its small increase reflects the shared schemas' module lifetime. No leak is
observed. Retain/release counts are not the same as recursive free work, so their
absolute difference is not a leak measure.

[gaps/arenaNodeRead.ts](gaps/arenaNodeRead.ts) isolates the owning-getter cost on
a file-created arena entry. 2.4 million reads produce Node's identical checksum
57,600,000,000 through either form. Counts are:

```text
accessor: allocations 10 frees 10 retains 4800011 releases 4800018 peak 10 regions 0
direct:   allocations 10 frees 10 retains 2400011 releases 2400018 peak 10 regions 0
```

Observed: the accessor adds one retain/release pair per read, with allocations
unchanged. This records a compiler ownership cost, not a correctness bug. The
matcher avoids the accessor locally; no compiler ownership pass was changed.

Validation and count commands:

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=30m ./stage1/cohere/yaml > /tmp/stage1-yaml-speed/final-suite.log 2>&1
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=10m ./stage1/cohere/yaml -run TestSpeedCostProbes > /tmp/stage1-yaml-speed/cost-probe-suite.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=15m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|collections|exceptions|bitwise)\.a$' > /tmp/stage1-yaml-speed/oracle.log 2>&1
go run ./cmd/adamic build stage1/cohere/yaml/gaps/arenaNodeRead.ts -o /tmp/stage1-yaml-speed/arena-read --count
/tmp/stage1-yaml-speed/arena-read /tmp/stage1-yaml-speed/unit-scan.txt accessor 100
/tmp/stage1-yaml-speed/arena-read /tmp/stage1-yaml-speed/unit-scan.txt direct 100
```

Logs: [full suite](audit/speed-fast-path-suite.log),
[cost probes](audit/speed-cost-probes.log), [oracle](audit/speed-oracle.log),
[lint](audit/speed-fast-path-lint.log), [format](audit/speed-fast-path-format.log).

## Direct emoji indexing and numeric-map cost

The emoji entry table now uses 65,536 UTF-16 slots rather than numeric Map keys.
Mapped surrogate units are converted back to their UTF-16 slot at construction;
the original matcher still uses the mapped rune representation. The table retains
branch order and supports multiple candidates if the pinned graph changes.
This trades a 64K reference-slot table (512 KiB of slots on this 64-bit build)
for direct lookup; RSS was not measured.

The post-matcher Map profile attributed 0.8 of 10.3 weighted seconds (7.77%) to
`find`, called by `adamic_map_get` under `Layout_text` (the inlined width path).
The direct-index profile has 9.5 weighted seconds and no sampled `find` cost.
These short profiles identify costs; the unprofiled benchmark decides elapsed
performance. Compiler/runtime files remain unchanged.

[gaps/numericMapLookup.ts](gaps/numericMapLookup.ts) isolates the runtime numeric
hash. It inserts 185 values, looks each up 10,000 times and prints the same
checksum 170,200,000 for integer or fractional key distributions. Five fresh
runs, with exact Node checksum and empty stderr, measured:

| Driver | Integer keys | Fractional keys |
| --- | ---: | ---: |
| Native release | 0.481702s | 0.031763s |
| Source Node | 0.087917s | 0.117898s |

Native integer lookups take 15.2 times fractional lookups here. Observation:
`map.c` hashes number bits with `(bits ^ (bits >> 29)) * 14695981039346656037`.
All integer keys 0..184 select the same initial bucket when masked to 512 buckets.
The probe and the sampler support hash collisions as the cause; no runtime hash
implementation was changed. This is a performance gap, not a correctness bug.
[Full samples](audit/speed-map-probe-timing.log) include every duration.

```sh
go run ./cmd/adamic build stage1/cohere/yaml/gaps/numericMapLookup.ts -o /tmp/stage1-yaml-speed/map-probe
/tmp/stage1-yaml-speed/map-probe 185 integer 10000
/tmp/stage1-yaml-speed/map-probe 185 fractional 10000
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/yaml/gaps/numericMapLookup.ts 185 integer 10000
```

The formatter/driver/mutant/Unicode-width/probe suite passed in 148.915s.
Every output remains Go-identical; all three cost probes also agree on sanitized
native, source Node and emitted JavaScript. Cohere's 276-rule check and format
check pass; vet and gofmt are empty. A fourth speed mutant shifts the surrogate
slot by 256: native and Node both finish successfully, and byte comparison alone
catches it at byte 865030. There are now 28 qualifying port mutants total.

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=20m ./stage1/cohere/yaml -run 'TestWidthsMatchGo|TestFormatter|TestFileDriver|TestSpeedCostProbes' > /tmp/stage1-yaml-speed/indexed-suite.log 2>&1
```

Logs: [direct-index suite](audit/speed-indexed-suite.log),
[profile](audit/speed-indexed-profile.log), [header](audit/speed-indexed-profile-header.log),
[lint](audit/speed-indexed-lint.log), [format](audit/speed-indexed-format.log).

## Reused lexer ASCII units

The lexer now builds its 128 ASCII unit strings once and reuses them for
character reads. Non-ASCII reads retain the original UTF-16 slice behavior;
out-of-range and NaN indices still return an empty string. This avoids the
one-unit allocations demonstrated by `stringUnitScan.ts`, without changing
the compiler or runtime. The table includes NUL and DEL as ordinary entries.

A fifth speed mutant replaces the cached NUL string with a space. Native and
source Node finish successfully with empty stderr; comparison catches the
wrong answer at byte 864537. There are 29 qualifying port mutants total.

The complete YAML suite passed in 316.275s. Lint then rejected the deliberate
`index !== index` NaN check. It was replaced with `Number.isNaN(index)` and the
lexer, formatter and direct driver checks were rerun, along with lint, formatting,
vet and gofmt. The final validation log is linked below.

The final counted build reproduces the 992,282 Go answer bytes and reports:

```text
allocations 5748774 frees 5748774 retains 28068956 releases 26155251 peak 67023 regions 0
```

That is 88.7% fewer allocations than the original 50,991,977. The cache trades
additional references for fewer allocations; retain counts are not reduced by
every individual optimization. These are object counts, not bytes or RSS.

Logs: [full suite](audit/speed-ascii-suite.log),
[final validation](audit/speed-ascii-final-suite.log),
[lint](audit/speed-ascii-lint.log), [format](audit/speed-ascii-format.log),
[counts](audit/speed-final-counts.log).

## Numeric lexer scans

Seven lexer paths now read numeric UTF-16 units: plain scalars, block scalars,
block scalar headers, indentation continuation, spaces, newlines and line-end
checks. Missing units use -1, distinct from a real NUL (0); whitespace and flow
indicator predicates preserve exactly the original sets. Token slices and all
state transitions are unchanged. Paths needing strings retain the ASCII cache.

An attempted `case -1` was refused by stage 0 as a nonconstant case.
[gaps/negativeCase.ts](gaps/negativeCase.ts) proves that refusal while Node prints
1. The port uses an explicit predicate instead; no lowerer change is needed.
This brings compiler refusal programs to eleven, separately from the three
performance probes and the existing shared-slice runtime correctness gap.

String-based tests delegate to that same numeric whitespace predicate, keeping
all seven paths consistent. The old string-only tab mutant ran for over 79s
without finishing and was stopped; it is not credited as successful wrong output.
It is replaced by changing the shared predicate from tab (9) to backspace (8).
The new numeric-whitespace mutant changes space (32) to unit 31. Both mutations
finish successfully with empty stderr on native and Node; comparisons catch tab
at byte 864210 and space at byte 6772. The cache mutant still qualifies (now
caught at byte 926523). There are 30 qualifying port mutants, six added in this
speed unit; the tab mutant replaces its former string-based definition.


The first numeric suite was stopped during the old tab mutant. It also exposed
a type-check error in the first negative-case probe (`console.log(number)`).
The probe now explicitly converts the result with `String`, reaches the intended
lowering refusal, and passes its separate 0.115s test. A complete suite is rerun
with both corrections; the interrupted log is retained for transparency.


The final counted native build matches Go and reports:

```text
allocations 5611765 frees 5611765 retains 25357452 releases 23306738 peak 67023 regions 0
```

Allocations fell 89.0% and retains 68.0% against the original. Compiler/runtime
files remain unchanged.

The complete suite passed in 306.891s, including all 30 qualifying mutants,
all external parser/printer oracles, eleven refusal programs, the separately
held runtime correctness gap, three cost probes and exhaustive Unicode width.
Lint passed (276 rules, 44 files); formatting requested only joining the block
header's two-line `else if` onto one line. That whitespace edit was applied and
all five lexer mutants were rerun. Final lint, format, vet and gofmt checks pass.

```sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=30m ./stage1/cohere/yaml > /tmp/stage1-yaml-speed/numeric-final-suite.log 2>&1
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=10m ./stage1/cohere/yaml -run '^TestLexerMutants$' > /tmp/stage1-yaml-speed/numeric-final-mutants.log 2>&1
```

Logs: [full suite](audit/speed-numeric-suite.log),
[final lexer mutants](audit/speed-numeric-mutants.log),
[interrupted attempt](audit/speed-numeric-attempt.log),
[negative-case probe](audit/speed-negative-case.log),
[lint](audit/speed-numeric-lint.log), [format](audit/speed-numeric-format.log),
[counts](audit/speed-numeric-counts.log).

## Reproducing the final throughput comparison

[benchmark_speed.py](benchmark_speed.py) rotates seven driver commands through
five fresh-process rounds. It times process startup, input, parsing, printing and
output, then checks every byte against Go and requires empty stderr and exit zero.
It uses blocking process waits rather than timeout polling in the timed interval.
No tests or profilers run concurrently with the final comparison. This is the
mixed successful/error edge corpus described in PERFORMANCE.md, not a production
workload claim. Source Node means this same port, not published Prettier.

To rebuild the before/after snapshots without changing the working branch:

```sh
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/stage1-yaml-speed
ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-format-artifacts ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library go test -v -count=1 -timeout=15m ./stage1/cohere/yaml -run '^TestFormatterMatchesGo$' > /tmp/stage1-yaml-speed/rebuild-corpus.log 2>&1
for pair in '77327e3 baseline' '26c379e shared-schema' '6385e9e fast-path-maps'; do
    set -- $pair
    mkdir -p /tmp/stage1-yaml-speed/$2-source
    git archive --format=tar --output=/tmp/stage1-yaml-speed/$2.tar $1 stage1/cohere/yaml
    tar -xf /tmp/stage1-yaml-speed/$2.tar --strip-components=3 -C /tmp/stage1-yaml-speed/$2-source
    go run ./cmd/adamic build /tmp/stage1-yaml-speed/$2-source/main.ts -o /tmp/stage1-yaml-speed/$2 > /tmp/stage1-yaml-speed/$2-build.log 2>&1
done
go run ./cmd/adamic build stage1/cohere/yaml/main.ts -o /tmp/stage1-yaml-speed/final > /tmp/stage1-yaml-speed/final-release-build.log 2>&1
python3 stage1/cohere/yaml/benchmark_speed.py /tmp/stage1-yaml-format-artifacts /tmp/stage1-yaml-speed --rounds 5 > /tmp/stage1-yaml-speed/final-timing.log 2>&1
```

The scratch library installation is pinned to yaml 2.9.0, yaml-unist-parser 3.2.0
and Prettier 3.9.6 as documented in GAPS.md. Toolchain setup timings remain in
PERFORMANCE.md and audit/setup.log: Go 0s, clang 1s, Node 1s, submodules 1s,
build cache 78s, total 78s; `nproc` reports 5, with a four-CPU quota.

## Final five-round results

Final implementation checkpoint: `f053bec`. Five interleaved rounds, with
no concurrent tests or profiles, checked every saved Go byte on all seven
drivers. All 35 runs exited zero with empty stderr and reproduced 992,282
answer bytes for 10,026 formatting calls (all 36 repository files included).

| Driver / checkpoint | Median texts/s | Median elapsed seconds |
| --- | ---: | ---: |
| Original native / 77327e3 | 2,808.75 | 3.569556 |
| Shared schemas / 26c379e | 8,003.25 | 1.252741 |
| Matcher fast paths with Map / 6385e9e | 9,753.31 | 1.027959 |
| Final native / f053bec | 11,504.03 | 0.871521 |
| Original source Node / 77327e3 | 7,838.61 | 1.279053 |
| Final source Node / f053bec | 13,818.88 | 0.725529 |
| Go cohere / 715ba94 | 15,190.25 | 0.660028 |

Native is 4.10 times its rebuilt baseline
and 1.47 times the original Node port.
The original 7,300 texts/s Node target is cleared. The same optimizations also
help Node: final native still takes 20.1% longer
than final Node and 32.0% longer than Go. No parity
with the optimized Node port is claimed. Further runtime/compiler work has
minimal programs here rather than changes to files owned by other workers.
The elapsed deficit to Node fell from 2.79 times (rebuilt original) to 1.20 times.

All five samples and input/answer SHA-256 hashes are in the
[measurement record](audit/speed-measurements.txt);
[full timing log](audit/speed-final-timing.log) holds every verified run.
Earlier single observations and timeout-polling measurements are exploratory;
this blocking-wait five-round comparison is the final report.

Green implementation commits, each pushed before proceeding:
`26c379e` shared schemas; `6385e9e` matcher/printer fast paths and width oracle;
`1429a5e` direct emoji indexing and map probe; `d2081b4` ASCII cache;
`f053bec` numeric lexer scans and unified whitespace predicates.
The final report/measurement commit follows those checkpoints.

The complete YAML suite, sanitizer checks, external libraries, exhaustive width,
mutants, lint, format and vet passed. An uncached filtered compiler oracle passed
in 1.777s (27 native misses, 18 Node misses, zero hits). The complete repository
gate was not run. Coverage remains Go cohere's default options and the held
corpus; throughput is measured on mixed valid/error edge cases, with repeated
chunk variants counted as separate calls. RSS and production workloads were
not measured. The pre-existing shared-slice runtime correctness gap remains
separately proved and is not used by the formatter.
