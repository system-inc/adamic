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
