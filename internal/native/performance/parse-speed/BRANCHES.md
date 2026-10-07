Built: instruction-address release/destruction branch profiles, two rejected branch prototypes and pinned interleaved best-of-ten timing; production runtime unchanged.
Commits: base 2cdf8dd; the isolated old release body is from 5ca9162, before the already pushed release fix 3196034.
Commands and outputs: four native profiles and Go reconcile all five events; 77 prototype trees match Go; deep-chain/shared/final/immortal release controls pass under sanitizers.
Mutants: omitted final destruction fails live-count checking and LeakSanitizer; one Bcm footer increment fails profile reconciliation; wrong timing stdout fails output checking.
Not covered: hardware branch/cycle events, a production change, implemented per-file regions, full native/oracle/repository packages in this unit, or the dated parse targets.

# Release branch shapes, October 7, 2026

The earlier instruction-only optimization is valuable, but simulated misses do
not move in the same direction as its instruction count. The existing outlined
release remains the winner in this timing series. Neither new prototype is
promoted. No runtime, string, emitter, parser, scanner or write-path code changes
are made in this unit.

## Exact build policy and inputs

Every native instruction/miss number and native timing below uses clang 20.1.8,
-O2 -g, sanitizers off, -ffp-contract=off, -fno-optimize-sibling-calls, no LTO and
no ADAMIC_COUNT. The complete build command, substituting MODE only, is:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I /workspace/scratch/parse-branches/MODE-runtime /workspace/scratch/parse-speed/runtime-parse.c /workspace/scratch/parse-branches/MODE-runtime/*.c -lm -o /workspace/scratch/parse-branches/MODE > /tmp/parse-branches-MODE-clang.log 2>&1
```

All four use the identical runtime-parse.c originally generated after merge
3cbc640. The driver is the accepted batch8 parse-only reconstruction: 77 pinned
TypeScript 6.0.3 compiler files at 050880ce59e30b356b686bd3144efe24f875ebc8,
reading, parsing, parent map, line table and cleanup, with rule traversal removed.
Current runtime snapshots are copied from 2cdf8dd. Old changes only heap.c back
to 5ca9162; other current fixes remain identical. This isolates the release
boundary rather than comparing two entire historical branches. Hint and ordered
change only the current public release body, as in branch_evidence/prototypes.patch.gz.

Go is the previously reconstructed parse-alone oracle executable, built with
Go 1.27.1 through cohere's typescript-go shim. Its source reads each same file,
ParseSourceFile and ECMALineMap, and keeps the source tree alive through the map.
It is Go, so clang flags do not apply to that row. Timing sets GOMAXPROCS=1 on
all commands. Go profiling additionally sets GODEBUG=asyncpreemptoff=1;
timing keeps ordinary Go preemption. Historical Go instruction counts are not
substituted for the fresh reference profile.

Setup: Go/clang/Node/submodules ready at 0s; cache warm and total 68s. nproc=5,
cpu.max=400000 100000, memory 17.6 GB. Each toolchain shell sources
/workspace/adamic-tools/env.sh. Raw profiles/assembly are gzip-compressed only
to retain their original bytes, including Callgrind's trailing whitespace.

## Whole-process comparison

The exact native release flags above apply to every native row. Bcm and Bim are
Callgrind's simulated conditional and indirect misses, not hardware events.
Best user and wall times are selected independently from ten samples per row.

| Variant | Ir | Bcm | Bim | Bcm + Bim | Best user s | Best wall s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Old combined release/destruction | 7,279,052,184 | 49,754,508 | 3,393,515 | 53,148,023 | 0.818442 | 0.867814 |
| Current outlined final destruction | 6,460,129,551 | 52,715,595 | 3,402,789 | 56,118,384 | 0.794595 | 0.845601 |
| Last-reference unlikely hint | 6,460,129,551 | 52,715,595 | 3,402,789 | 56,118,384 | 0.801240 | 0.842017 |
| Shared-reference-first ordering | 6,470,113,506 | 51,714,257 | 3,402,789 | 55,117,046 | 0.816926 | 0.858016 |
| Go parse alone | 1,669,967,613 | 10,470,999 | 3,295,828 | 13,766,827 | 0.137863 | 0.180358 |

The existing outlined path saves 818,922,633 Ir and 2.91% best user time relative
to the matched old release. Whole-process simulated misses increase 2,970,361.
The instruction delta is eight smaller than the previous 818,922,637 comparison;
whole-process startup costs differ by a few instructions between measurement
runs. No attempt is made to hide that discrepancy. The current fresh Ir differs
from the earlier 6,460,129,543 by eight instructions.

The hint produces byte-identical adamic_release assembly and identical complete
profile events to current. Timing variation between these identical code shapes
is not evidence that the hint optimizes anything.

Shared-first cuts whole-process misses by 1,001,338 but adds 9,983,955 Ir and is
2.81% slower in best user time. Its public release conditions actually increase
misses by 54,769; changes elsewhere account for the whole-process decrease.
This is rejected, not a claimed speedup from a better branch shape.

Current native/Go user time is about 5.76x, versus 3.87x in fresh Ir. This supports
measuring beyond instructions, but does not prove that simulated predictor misses
cause that gap. User time is scheduled execution, not cycles or IPC. Simulator
address/history aliasing, memory traffic, real cache/predictor behavior, frequency
and execution overlap prevent a hardware causal attribution on this evidence.

## Which release conditions mispredict

Current addresses identify exact native branches under the release flags above.
They are attached to assembly and source records, not inferred from inclusive
function totals. Taken counts come directly from Callgrind's jcnd records.

| Public release condition | Address | Executions | Taken | Bcm |
| --- | --- | ---: | ---: | ---: |
| value is NULL | 0x72973 | 40,139,810 | 3,529 | 452,825 |
| header references is zero, immortal/region | 0x7297b | 40,136,281 | 16,386,567 | 5,122,564 |
| decrement reaches zero, last reference | 0x72983 | 23,749,714 | 960,535 | 1,002,671 |

The dominant public branch distinguishes immortal from counted values. It is
not the last-reference case. Shared counted releases are the common counted
path: 22,789,179 decrements remain nonzero. Only 4.04% of counted decrements
reach destruction. There is no write to immortal headers in any prototype.

The very rare NULL outcome still gets 452,825 simulated misses. Outcome
frequency therefore cannot be read as misprediction frequency: the simulator's
finite predictor uses branch addresses and history, shared with other code.
The same NULL branch grows to 1,094,669 misses in the ordered prototype although
its executions and outcomes are unchanged. This is direct evidence that a source
condition's rarity alone does not determine these simulated counts.

Inside destruction, the largest current branches are:

| Mechanism | Function/address | Executions | Misses |
| --- | --- | ---: | ---: |
| Runtime field type says reference or scalar | object_free_children 0x6e4b8 | 13,679,516 conditional | 4,223,986 Bcm |
| Child header has a count to drop | object_free_children 0x6e4c8 | 6,299,752 conditional | 1,866,476 Bcm |
| Class field-loop continuation | object_free_children 0x6e4a3 | 14,762,530 conditional | 1,530,621 Bcm |
| Heap kind switch jump table | destroy_last_reference 0x72a85 | 3,377,171 indirect | 2,714,786 Bim |
| Array property child is NULL | destroy_last_reference 0x72d23 | 1,557,564 conditional | 844,805 Bcm |

The array-property branch is the NULL guard in inlined let_go, before any
header read or decrement, not a last-reference test or allocation/free syscall.
The child-header and final-decrement tests are separate instructions.
The kind switch performs an indirect dispatch per destroyed object. Field
metadata is inspected per slot in the derived/base field walk. These confirm
where generic destruction pays control-flow costs; ordering the external release
does not remove them. The existing iterative queue remains bounded in C stack
depth and is unchanged by both prototypes.

For a reproducible named-path subtotal, sum only adamic_release,
destroy_last_reference, adamic_object_free_children, let_go and
adamic_map_free_children self events. Old has 1,514,870,246 Ir and 17,645,832
misses; current has 695,947,609 Ir and 18,874,615 misses; ordered has 705,931,564 Ir
and 18,605,562 misses. These are disjoint functions, not the earlier report's
mechanism bucket. Inline slab code is included in their containing functions.
The cited release-lto report's 16.597M mechanism-bucket misses use runtime base
1740da37 and different inline attribution/layout. They must not be equated with
these named-function sums on 2cdf8dd.

Class-specialized destruction could remove the measured per-field type branch
and loop, while a proven one-stroke per-file region could remove much structural
teardown. Those remain proposals requiring ownership/escape proof and, for
specialization, a small compiler hook cleared with the compiler owner. No such
hook is edited or pushed here. DESTRUCTION.md's lifetime/count and per-file
region estimates remain estimates; this unit does not implement a region or
turn simulated misses into a claimed region time saving.

## Measurements and checks

```sh
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --branch-sim=yes --dump-instr=yes --collect-jumps=yes --callgrind-out-file=/workspace/scratch/parse-branches/MODE.callgrind /workspace/scratch/parse-branches/MODE --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/parse-branches/MODE.stdout 2> /workspace/scratch/parse-branches/MODE.stderr
python3 internal/native/performance/parse-speed/branch_profile.py /workspace/scratch/parse-branches/MODE.callgrind --output /workspace/scratch/parse-branches/MODE.json
python3 internal/native/performance/parse-speed/branch_timing.py /workspace/scratch/parse-branches --go /workspace/scratch/parse-speed/go-parse --manifest /workspace/scratch/parse-speed/compiler.txt --cpu 3 > /workspace/scratch/parse-branches/timing.log 2>&1
```

branch_profile.py handles instruction/line deltas, inline file and jump symbols,
and excludes inclusive call-edge vectors. All five native and Go self-event sums
match both totals footer and summary exactly. It retains per-instruction Bc/Bcm,
Bi/Bim and conditional taken counts, so every reported condition is inspectable.

branch_timing.py warms each command once, then rotates and reverses the five
programs across ten interleaved rounds. Each child is pinned to CPU 3 before
exec; wait4 captures that child's user/system time directly. Every ordinary
stdout must be exactly 0 followed by newline, every stderr empty, exit zero.
All 50 measured captures passed. No compiler, test, profiler or other benchmark
ran concurrently during timing. Background container services remain. One-minute
load was 0.78 to 0.89; host isolation and frequency are not proven. Raw commands,
all samples and load observations are in branch_evidence/timing.json.

The ordered prototype's separate whole-tree driver matches independent Go on
all 77 files, 44,766,682 bytes under ASan/UBSan/LeakSanitizer: scratch overlay
TestReleaseBranchPrototypeParity PASS 10.494s. Production
TestRuntimeReleasePaths PASS 0.310s. The scratch counted sanitizer fixture also
checks NULL, immortal, shared, final string destruction and a 100,000-object
chain with runtime-built labels: 200,001 allocations and 200,001 frees, no leaks.
These correctness builds use the release warning/runtime flags above but
-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; the counted fixture
also uses -DADAMIC_COUNT. They are not the profiled/timed binaries.

Omitting the ordered prototype's final helper call compiles and runs but fails
`last release left 1 values`, then LeakSanitizer reports a 78-byte leak. Changing
only the profile footer Bcm by one raises self/footer/summary mismatch. A fresh
child producing stdout 1 instead of 0 exits cleanly but fails the timing output
checker. These mutants are scratch-only and do not replace measured good binaries.
The latter two prove the new measurement checks can fail independently of
program correctness. Vet of the parser/native packages passes with empty output.
No full native/oracle package gate is newly claimed for this evidence-only unit;
the unchanged production code retains the prior completed gates in FIELDS.md.

Raw profiles, assembly, parsed event vectors, prototype diffs, hashes, mutants,
parity and release logs are in branch_evidence. Scripts and these artifacts are
the only repository changes. Continue reporting all three of Ir, simulated
misses and pinned interleaved best-of-ten user time for subsequent release work.
