Built: cache and ordinary user-time comparisons for the scratch forced-inline change and a clang-selected alternative; no production change.
Commits: retained binaries from c01907a; model/reference 968a7cc; current main c7991b9 merged as b5ca630.
Commands and outputs: all five cache profiles and all 24 ordinary timing invocations exit 0 and match expected stdout.
Mutants: each totals-footer event +1 is rejected, 13 events across five profiles, 65 actual corrupt profiles caught.
Not covered: hardware cache counters/cycles, a proven timing speedup, production check removal, or the original missing headroom proof.

## Instructions, instruction-cache misses and user time

These measurements use the retained c01907a binaries from the first report,
not binaries rebuilt against the newer main merged for delivery. The original
SPLIT.md driver, 77-file corpus, release flags and runtime remain fixed. See the
parent REPORT.md for their exact preparation and compiler flags.

The extra scratch variant, `natural`, removes the same three stack checks but
adds no inline attribute. The original `after` variant additionally forces all
three helpers inline. No emitter or runtime code was changed for these runs.

| Parse, whole process | Baseline | No three checks, clang chooses inlining | No three checks, always_inline |
|---|---:|---:|---:|
| Cache-run instructions, Ir | 9,259,808,338 | 8,985,321,537 | 8,738,361,858 |
| Ir change | reference | -2.9643% | -5.6313% |
| Modeled L1 instruction misses, I1mr | 62,460,653 | 58,662,271 | 62,575,742 |
| I1mr change | reference | -6.0812% | +0.1843% |
| Median ordinary user time, ms | 1,085.949 | 1,080.021 | 1,113.369 |
| Best ordinary user time, ms | 1,015.476 | 971.343 | 1,000.871 |
| Executable .text bytes | 489,582 | 491,086 | 525,166 |
| .text change | reference | +0.3072% | +7.2672% |

The instruction advantage of forced inlining does not establish a timing win:
its median user time is 2.53% above baseline in this run. The clang-selected
variant has a 0.55% lower median user time, inside the noise of this shared
machine. Its smaller code expansion and fewer modeled instruction misses make
it a better starting point for a future sound placement implementation than a
blanket always_inline policy. Neither variant is production-ready: both remove
checks by hand in scratch, and neither supplies the headroom proof.

Both scratch variants run the same parse-only count driver and print `0\n`.
This validates that measured invocation, not full AST parity. All ordinary
runs have empty stderr; cache stderr retains the simulator's warnings.

The explicitly matching emitter-speed model is:

```text
--tool=callgrind --cache-sim=yes --branch-sim=yes
--I1=32768,8,64 --D1=32768,8,64 --LL=268435456,1,64
```

I1 and D1 are 32 KiB, eight ways, 64-byte lines; LL is 256 MiB, direct mapped,
64-byte lines. Callgrind 3.24.0 models each process independently. This is the
same stated model as 968a7cc, not the physical machine's cache hierarchy, and
not a simulation of cross-process cache contention. Parse profiling processes
ran concurrently. No elapsed or user time under Valgrind is used as a native
timing result.

The reference report's 55,805,810 L1 misses and 48,840,075 instruction misses
are its **remainder source bucket**, not its complete parse total. Its separate
whole-process baseline is 62,465,543 I1 misses. This report uses its own
whole-process pair throughout; no different-runtime baseline is substituted.

Ordinary timing uses the original Workers wait4 helper's direct-child ru_utime,
with one warm invocation per variant, then seven complete parse runs per
variant. Order rotates before/natural/after by round. No profiling, compilation
or test gate overlaps the ordinary timing block. No startup subtraction is
applied to parse: both the instruction profile and user time cover the whole
process. [user-time.json](user-time.json) retains every measured sample, user,
system and wall times, load averages, commands and binary hashes. User time is
reported independently of user-plus-system CPU time.

.text is the ELF section's Size field from `objdump -h`. The model data cannot
prove that the size change causes a particular miss change; layout, conflict
and control-flow changes were not isolated. All other simulated events,
including data misses and branch mispredictions, are in [metrics.json](metrics.json).

## Workers sieve

The prior before/after sieve binaries are byte-identical and remain unchanged.
New cache runs use its original three-request corpus with K=1. User-time rows
are derived from the prior three ordinary paired K=1000/K=0 rounds, using
`(measured.userMs - startup.userMs) / 3000`, not combined CPU time. These are
explicitly different profiling and timing batch sizes.

| Native sieve | Before | Inactive classifier |
|---|---:|---:|
| Whole-process cache-run Ir, K=1 | 20,963,517 | 20,963,503 |
| Whole-process modeled I1mr, K=1 | 1,802 | 1,802 |
| Median startup-subtracted user ms/request, K=1000 | 0.723948 | 0.718666 |

The tiny profile differences and timing differences establish no compiler
improvement: the executable bytes are identical. Original raw user-time samples
remain in ../sieve-measurements.json; new cache profiles are retained here.
The original standard -O2 native flags still apply, without sanitizers,
counting, -g or LTO for sieve. The native-only Workers coverage limitation
from the parent report remains.

## Reconciliation and reproduction

Each of 13 self-event sums equals its totals footer. Every profile summary has
Ir exactly 2 above the self/footer total, and agrees on all other events. Tables
use self/footer values. These cache-run instruction totals are recorded
separately from the earlier Ir-only run; the small difference is not silently
assigned to any mechanism.

summarize.py is the unchanged Adamic helper from 968a7cc. reconcile.py additionally
writes a corrupted raw profile for each footer event, invokes that summarizer,
and requires `self events differ from totals footer`. All 65 are caught;
[event-mutants.json](event-mutants.json) records the original and corrupted value
and actual failure. The original compiler mutants from the parent report were
not rerun for this measurement-only update.

Example cache invocation (repeat for native-natural and native-after):

```sh
VALGRIND_LIB=/workspace/scratch/stack-check-scc/valgrind/usr/libexec/valgrind /workspace/scratch/stack-check-scc/valgrind/usr/bin/valgrind --tool=callgrind --cache-sim=yes --branch-sim=yes --I1=32768,8,64 --D1=32768,8,64 --LL=268435456,1,64 --callgrind-out-file=/workspace/scratch/stack-check-scc/cache/parse-before.callgrind /workspace/scratch/stack-check-scc/native-before --manifest /workspace/scratch/stack-check-scc/compiler.txt --count > /workspace/scratch/stack-check-scc/cache/parse-before.stdout 2> /workspace/scratch/stack-check-scc/cache/parse-before.stderr
```

For sieve, substitute handler/native-before or handler/native-after and pass
`/workspace/scratch/stack-check-scc/handler/sieve.txt 1` instead of manifest flags.
The unchanged natural C is archived as parse-natural.c.gz; build it using the
parent's exact clang line with parse-natural.c and native-natural substituted.

After all profiling processes have completed:

```sh
source /workspace/adamic-tools/env.sh
python3 /workspace/scratch/stack-check-scc/cache/reconcile.py > /workspace/scratch/stack-check-scc/cache/reconcile.log 2>&1
python3 /workspace/scratch/stack-check-scc/cache/user_time.py > /workspace/scratch/stack-check-scc/cache/user-time.log 2>&1
```

Future emitter comparisons in this unit must report Ir, I1mr and ordinary user
time together with model/build flags. Prefer compact hot code and cold failure
paths out of line when their measured effects support that choice. A hot-loop
call or a small body is a reason to test inlining, not evidence that it wins.
The other emitter workers' switch/store/name territory remains untouched.

## Landing checks

Current origin/main c7991b9 merged cleanly as b5ca630. The uncached native and
oracle package gate passes on that merged branch: native 158.602s, oracle
147.103s, including the existing sanitizer lanes. Touched-package vet and
`git diff --check` pass. This is not the full repository gate. Compiler code
was not changed in this follow-up; the benchmark artifacts remain pinned to
c01907a, not the merged checkout.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/stack-check-metrics-native-oracle.log 2>&1
go vet ./internal/native > /tmp/stack-check-metrics-vet.log 2>&1
```
