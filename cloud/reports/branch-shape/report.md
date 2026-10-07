Built: guarded numeric-switch dispatch for review; scalar lowering rejected; no repeatable user-time speedup demonstrated.
Commits: fix 3f8603c; claims 654d40f and 696e466; merged current main c7991b9 in ad50a03. This report and evidence are committed afterward.
Commands and outputs: clang 20.1.8 -O2 -g, Callgrind branch/cache simulation and pinned best-of-ten user time; merged native/lower/uncached oracle PASS; batch 8 Go parity PASS.
Mutants: fractional conversion caught by Node output; missing range guard caught by UBSan; rejected eager-call prototype caught by Node output; corrupt profile footer and timing stdout rejected.
Not covered: a validated fix for the largest short-circuit branch source, tag-check elision, string-switch acceleration, hardware counters, full repository gate, or other platforms.

# Branch shape in emitted C

The accepted change is in `internal/native/emit_statements.go` and the new
`internal/native/branch_shape.go`, with `branch_shape_test.go` holding its semantics.
No production changes remain in emit_expressions.go. No field stores, sieve,
data layout, stack-check placement, symbol counters, integer-field emission,
cohere source, or forbidden compiler files were edited by this unit.
The runtime and lowering changes brought in by the required main merge belong
to main, not this unit.

## Inputs, tools and flags

Started from current main 39638d9, on codex/branch-shape. Fetched the supplied
reports at parse-speed a709d575f574ddf8dc47460873a07b4d5fc4b234 and release-lto
c7c8c0bb74c609baaa66814221653419cbc1bddb. Later merged main
c7991b900362796aefd111474e65eb5398e91953 and re-greened it.

The unmodified parse-speed prepare.py reconstructs batch 8 from
4189abd3490757e8abe13722ceb365c451293e92. Its scratch .a parse driver removes
only `visit(context, root)`. Parser, scanner, Context, parent map, line table,
sorting, output and destruction remain. The invocation is
`parse --manifest compiler.txt --count`. TypeScript is pinned at
050880ce59e30b356b686bd3144efe24f875ebc8, exactly 77 compiler files.
The corpus SHA-256 manifest is retained. cohere is
715ba94f3608a6500086b1076ce5cb7e51b836db, accessed through its existing Go shims.
No cohere implementation was copied.

`bash cloud/setup.sh` passed: Go ready 0s; clang ready 1s; Node ready 1s;
submodules ready 1s; build cache warm 115s; done 115s.
Every build/test shell sourced /workspace/adamic-tools/env.sh.
Go 1.27.1, Node 24.19.0, clang 20.1.8, Valgrind 3.24.0, Linux 6.18.44,
AMD EPYC 9V74, nproc=5, cpu.max=400000 100000. Valgrind was absent, so
valgrind_3.24.0-3_amd64.deb was extracted in scratch from
https://deb.debian.org/debian/pool/main/v/valgrind/valgrind_3.24.0-3_amd64.deb.
The initial harness preparation needed its historical 4189abd commit fetched;
it then prepared all 77 files successfully.

All profiled and timed native binaries use precisely:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic \
  -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function \
  -Wno-unused-parameter -Wno-self-assign -ffp-contract=off \
  -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime \
  /workspace/scratch/branch-shape/VARIANT.c internal/native/runtime/*.c \
  -lm -o /workspace/scratch/branch-shape/VARIANT
```

No LTO, PGO, sanitizers, ADAMIC_COUNT, ADAMIC_SLABS or CPU-specific flags.
Each runtime .c is a separate translation unit. Timing uses the same debug
executable as simulation. The separate batch 8 correctness release uses the
ordinary native.Build -O2 policy; sanitized pinned tests also run.

```sh
VALGRIND_LIB=/workspace/scratch/branch-shape/valgrind/usr/libexec/valgrind \
/workspace/scratch/branch-shape/valgrind/usr/bin/valgrind \
  --tool=callgrind --cache-sim=yes --branch-sim=yes \
  --dump-instr=yes --collect-jumps=yes \
  --callgrind-out-file=/workspace/scratch/branch-shape/VARIANT.callgrind \
  /workspace/scratch/branch-shape/VARIANT \
  --manifest /workspace/scratch/branch-shape/compiler.txt --count \
  > /workspace/scratch/branch-shape/VARIANT.stdout \
  2> /workspace/scratch/branch-shape/VARIANT.stderr
```

I1 and D1 are 32KiB, 64-byte lines, 8-way; LL is 256MiB, direct mapped.
These are simulator events, not the EPYC's measured cache/predictor behavior.
Callgrind reports its LL-model and nonfatal brk-segment warnings, retained in
stderr. All real profiles finish successfully with stdout exactly `0\n`.
Self costs exclude inclusive call edges and reconcile all 13 events with each
footer. Each summary has exactly 2 more Ir than its footer and self sum, with
no other event discrepancy. That difference is preserved rather than allocated
to any function. Reported instructions are the reconciled self/footer totals.

Timing warms each executable once. The four-variant prototype series rotates
and reverses its order; the final two-variant series explicitly alternates
which executable runs first across ten interleaved rounds. Each child is pinned to CPU 3 before exec;
os.wait4 reports that child's user time. Every capture checks exact stdout,
empty stderr and exit zero. Builds, tests, profilers and other benchmarks were
finished before timing. Raw times, orders, commands and load observations remain
in timing.json and timing-landing.json. Host-wide isolation and frequency are
not certified. The final 1.20% user-time increase is small and may include noise. Earlier
series favored the switch, so a repeatable user-time improvement is not established.

## Generated-function ranking

This is the fresh post-merge no-switch control, not an imported historical
ranking. Sum Bcm + Bim as self events, and combine Callgrind's recursive
contexts (`function`, `function'2`, etc.) into the same generated C function.
Runtime functions are excluded from this generated-function table. Inline
helpers contribute to their containing generated caller. Evidence retains the
unaggregated profile as well as aggregated vectors, emitted C and machine code.
C line numbers below refer to landing-before.c, which is byte-identical to before.c.

| Rank | Generated function | Simulated misses | L1 instruction misses | Costing shape and C/assembly evidence |
| --- | --- | ---: | ---: | --- |
| 1 | adamic_function_33_Scanner_scan | 2,439,629 | 3,218,907 | Numeric short-circuit predicates, including inlined isSpace. 3694: ucomisd / jne chain; its largest branch has about 1.06M misses. |
| 2 | adamic_function_1_isIdentifierStart | 1,511,488 | 446,121 | ASCII equality and range short circuits. 3642: ucomisd / jb / jae; the first two range branches account for about 1.32M misses. |
| 3 | adamic_function_62_Context_mapParents | 903,942 | 308 | Empty child-list guard and loop backedge. 9482: cmp / je before the loop, cmp / jb at its bottom; already rotated. |
| 4 | adamic_function_34_Scanner_punctuation | 782,726 | 2,688,918 | Numeric character switch emitted as an equality chain. 8041/8045: ucomisd / jne / jp, before the fix; switch now has guarded conversion and jmp *rax. |
| 5 | adamic_function_11_precedence | 689,417 | 1,367,167 | String token switch, with grouped || tests. 4038/4053: string_equal calls followed by test / conditional jumps. |
| 6 | adamic_function_153_Statements_statement | 390,824 | 6,515,260 | Statement-kind dispatch and method/field fallbacks. 20905/20540: conditional dispatch; 21213: shape/class/cache branches and method calls. |
| 7 | adamic_function_92_Parser_type | 305,227 | 1,532,145 | Type grammar conditionals and short-circuit token checks. 11084/11159/11874: conditional branches around string tests and calls. |
| 8 | adamic_function_122_Parser_unary | 252,737 | 2,091,181 | Unary token membership short circuits. 15322: string_equal calls and conditional jumps across six token tests. |
| 9 | adamic_function_81_Parser_primary | 250,175 | 2,122,564 | String token switch emitted as an equality chain. 10345/10314: string_equal calls and conditional jumps across token cases. |
| 10 | adamic_function_73_Parser_kind | 225,408 | 2,984 | Stack-limit guard; field shape fallbacks are smaller. 9853: cmp stack limit / ja; 9854: field shape comparison. |

The top ten's emitted C and clang-produced machine code are in
[evidence/top10.json](evidence/top10.json), one .txt per function, and the complete
compressed generated C. The hottest branches were inspected at their measured
instruction addresses, including inlined isSpace inside Scanner_scan.
The parent-map assembly has an empty-array entry guard and a rotated backedge;
it does not test the top condition a second time on every iteration.
No repeated union-tag test with a valid elision proof was identified in this
ranking. Stack and field guards stay with their owners. Outcome rarity alone
cannot justify dropping them or explain simulator misses.

## Largest candidate: eager scalar conditions, rejected

A closed whitelist of local scalar constants, unchecked noncaptured nonglobal
reads, comparisons and nested boolean operators allowed C &/| in place of
&&/||. Calls, property/index reads, unwraps and checked/global/captured reads
stayed short-circuiting. The prototype passed Node comparisons and the full
native/oracle packages, but its measured trade-offs were poor.

These comparisons share main 39638d9 and identical runtime/flags. scalar-only
is the logical prototype without integer-switch lowering, built through a Go
source overlay. scalar combines both changes.

| Version | Bcm + Bim | L1 instruction misses | L1 data misses | All L1 misses | Instructions | Best of 10 user s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| before | 45,542,108 | 62,459,761 | 22,515,164 | 84,974,925 | 9,259,839,513 | 1.001991 |
| scalar-only | 51,423,731 | 64,566,655 | 22,845,731 | 87,412,386 | 9,258,360,676 | 0.990636 |

The standalone prototype increases instruction misses 3.37% and simulated
misses 12.91%; best user time improves only 1.13%. Its isIdentifierStart body
grows from 200 to 240 bytes. Both Scanner_scan and isIdentifierStart simulated
self misses increase. Fewer conditional executions did not establish a win.

The incremental experiment after the switch fix is:

| Version | Bcm + Bim | L1 instruction misses | L1 data misses | All L1 misses | Instructions | Best of 10 user s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| switch | 45,708,345 | 55,682,953 | 22,368,314 | 78,051,267 | 9,248,336,226 | 0.965419 |
| scalar | 47,027,220 | 54,010,770 | 22,162,581 | 76,173,351 | 9,246,776,218 | 1.007126 |

Adding it slows best user time 4.32% versus switches alone, despite reducing
instruction misses 3.00%. It is rejected and fully removed. Prototype hooks,
helper source, pinned Node test, raw profiles and mutants are preserved in
evidence. The largest short-circuit branch source is therefore still unresolved;
this unit does not claim that eliminating short circuits generally is impossible.

## Numeric switch dispatch, retained for review

Constant int32 numeric cases with at least four distinct labels emit a real C
switch. The switch value is still evaluated once. A range guard precedes the
int64 conversion and a round-trip equality guard prevents fractional inputs
from matching integer cases. Other inputs select a sentinel outside every
eligible label. NaN, infinities, huge values and fractions keep strict-equality
semantics. Duplicate labels keep the first case, including positive/negative
zero. Nonconstant/nonintegral/out-of-range cases and smaller switches retain
the existing comparison-chain lowering. Case scopes, cleanup, break and
continue behavior remain intact.

clang chooses a jump table for Scanner_punctuation. Its native body shrinks
from 2,191 to 2,119 bytes. In the original matched pair its self instructions
fall from 34,958,843 to 23,560,742; self instruction misses fall from 2,537,587
to 936,946; conditional plus indirect misses fall from 782,726 to 570,859.
The new indirect misses are reported, not hidden. Whole-process simulator
changes also depend on layout and predictor/cache aliasing; function byte size
alone does not explain the whole cache delta.

### Primary matched post-merge measurement

Both controls use main c7991b9's runtime/header sources. The final compiler's
emitted parse C matches the previously measured switch C byte for byte.
A no-switch Go overlay on the merged compiler emits C byte-identical to the
original baseline. Thus these are matched builds on current merged main.

| Version | Bcm + Bim | L1 instruction misses | L1 data misses | All L1 misses | Instructions | Best of 10 user s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| landing-before | 45,795,493 | 63,840,562 | 22,501,860 | 86,342,422 | 9,259,505,951 | 0.977822 |
| landing | 45,726,690 | 58,807,339 | 22,577,605 | 81,384,944 | 9,248,395,512 | 0.989602 |

Simulated misses fall 0.15%, instruction misses 7.88%, all L1 misses 5.74%,
and instructions 0.12%. Best user time increases 1.20%; data misses increase
by 75,745. This is a general code-shape and simulated-cache improvement, but it
is not a validated time improvement. The implementation stays on the review
branch, with no claim that the performance objective is complete.

### Earlier matched measurement, before main advanced

| Version | Bcm + Bim | L1 instruction misses | L1 data misses | All L1 misses | Instructions | Best of 10 user s |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| before | 45,542,108 | 62,459,761 | 22,515,164 | 84,974,925 | 9,259,839,513 | 1.001991 |
| switch | 45,708,345 | 55,682,953 | 22,368,314 | 78,051,267 | 9,248,336,226 | 0.965419 |

Here simulated misses increased 0.36%, instruction misses fell 10.85%, and best
user time improved 3.65%. All series are retained; they must not be blended.
The main merge introduced runtime/header code and changed the measured layout,
while both generated C variants stayed identical. This sensitivity limits causal
claims about simulated predictor counts. No result from emitter-speed's inline
stores is substituted for this control.

The initial post-merge two-variant series reported 0.980009s control and
0.965336s switch. Inspection found that rotating and then reversing two names
kept the control first in every pair. That series is retained as
timing-landing-initial.json, but is supplemental, not the primary comparison.
The corrected alternating series above is the primary result. It was not
repeated again in search of a favorable result.

## Correctness and mutants

The final merged tree passes:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m \
  ./internal/lower ./internal/native ./internal/oracle \
  > /workspace/scratch/branch-shape/landing-gate.log 2>&1
go vet ./... > /workspace/scratch/branch-shape/landing-vet.log 2>&1
gofmt -l cmd internal > /workspace/scratch/branch-shape/landing-gofmt.log
```

lower: 16.881s; native: 172.536s; oracle: 167.198s. Vet and formatting logs are
empty. Every ordinary oracle fixture is checked uncached against Node. No
counts table was edited by this unit. The earlier combined prototype also
passed native/oracle, 142.904s / 141.352s. A later gate crossed the main merge
and failed TestCountsAreRecorded without reporting changed row values; its
log is retained and it is not counted as final validation. The clean merged
gate above is the final result.

The full batch 8 release driver, including rule traversal, matches the independent
Go oracle on all 77 files: 11,444,034 bytes, empty stderr, exit 0. Both compressed
outputs and their hashes are retained. This is full diagnostic/fix output, not
just the parse count `0`. The Go oracle is built inside cohere with the existing
batch8_oracle.go overlay; no cohere files are modified. The merged compiler was
used for the final release comparison.

| Mutant actually compiled/run | Catcher and observed result |
| --- | --- |
| Remove numeric conversion's round-trip equality guard | TestNumericSwitchMatchesNode: release output maps 1.5 to one instead of other; exit 1 |
| Remove numeric conversion's range guard | TestNumericSwitchMatchesNode: sanitized binary reports UBSan, NaN outside representable long range; exit 1 |
| Allow eager user calls in rejected scalar prototype | TestScalarLogicalMatchesNode: Node counter 0/1 becomes native 2/3; release output mismatch, exit 1 |
| Increase profile footer Bcm by one | Reconciler raises self/footer/summary mismatch |
| Timing command exits 0 with stdout 1 instead of 0 | Observer raises output/exit mismatch |

The first two were rerun on the merged final compiler, each exit 1. No program
mutant was killed by a warning or compile-time refusal. The pinned numeric test
also covers zero/negative zero, int32 maximum, fractions, NaN, both infinities,
values outside the case range, duplicates, break, continue and owned case locals,
against source on Node, release native and sanitized native. No lowered runtime
check was elided; the new conversion guards are proven necessary by mutants.

## Reproduction and remaining scope

Raw profiles, event vectors, scripts, generated C, disassembly, complete timing
samples, runtime hashes for both main revisions, corpus hashes, compared batch 8
output and gate/mutant logs are under evidence/. Build and simulation commands
above use absolute scratch paths intentionally. The captured source overlay
replaces only emit_statements.go with its original implementation for the control.
The scalar prototype can be reconstructed from the retained hook diff, helper
source and test source; it is not production code.

To recreate the inputs, fetch 4189abd and the pinned TypeScript commit, obtain
prepare.py and testdata/go_parse.go from parse-speed a709d575, set its repo path
to the current Adamic checkout, and run it with the scratch and corpus directories.
Go batch 8 uses `go build -overlay=overlay.json` inside cohere on the virtual
cohere/adamic_batch8_oracle.go path; the overlay is generated by prepare.py.
Run each binary with the same absolute manifest paths. Runtime source hashes
and binary/generated-C SHA-256 values identify the exact measured artifacts.

Not completed: a successful general fix for the top scalar short circuits,
string-token switches, tag-check elision with an invalidation proof, or a new
loop transformation. Existing rotated loops were left alone. Hardware PMU
measurements, other platforms/LLVM versions and the full repository gate were
not run. The touched packages and entire uncached oracle were run instead,
as the worker gate permits. No pull request was opened, and no main or area
branch was pushed.
