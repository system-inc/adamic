Built: scratch parse comparison replacing only -fno-optimize-sibling-calls with -foptimize-sibling-calls; no compiler or runtime changes.
Commits: benchmark source/runtime c01907a; delivery follows 57d4ef1 on codex/stack-check-scc, with origin/main c7991b9 already merged.
Commands and outputs: both cache profiles and 22 ordinary runs exit 0, print 0, with ordinary stderr empty; results below.
Mutants: all 26 actual totals-footer +1 mutants rejected by profile self-event reconciliation.
Not covered: selective attributes, runtime callback changes, new recursion guards, or LTO; the under-2% gate stops this item.

## Decision

Stop at measurement, as requested. Enabling sibling-call optimization increases
instructions by 0.11764% and best-of-ten user time by 0.81614%, while reducing
modeled L1 instruction misses by 0.62848%. It establishes no 2% gain. The global
-fno-optimize-sibling-calls flag and -ffp-contract=off remain unchanged.
No disable_tail_calls attributes or runtime files were changed.

| Whole parse process | -fno-optimize-sibling-calls | -foptimize-sibling-calls | Change |
|---|---:|---:|---:|
| Instructions, Ir | 9,259,808,305 | 9,270,701,624 | +0.11764% |
| Modeled L1 instruction misses, I1mr | 62,460,652 | 62,068,101 | -0.62848% |
| Best of ten ordinary user time, ms | 996.152 | 1,004.282 | +0.81614% |
| Median of ten ordinary user time, ms | 1,032.1875 | 1,042.314 | +0.98107% |
| Executable .text bytes | 489,582 | 489,950 | +0.07517% |

## Inputs and flags

Both binaries were rebuilt from the original parse-before.c with all emitted
stack checks intact. Neither three-check removal nor always_inline attributes
from the earlier scratch experiments is present. The same SPLIT.md batch-8
parse-only driver and 77-file TypeScript 6.0.3 corpus are used; see the parent
REPORT.md for exact driver/corpus pins and hashes. Runtime C and headers were
extracted from c01907a with git archive into a scratch directory, so both sides
use the same runtime as that original benchmark rather than newer main's runtime.
This is a controlled flag comparison on that pinned workload, not a claim about
the newer main or a future SCC placement implementation. The original classifier
remains inactive pending the headroom proof.

Common clang 20.1.8 build flags:

```text
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -ffp-contract=off -O2 -g
```

The baseline adds -fno-optimize-sibling-calls; the experiment instead adds
-foptimize-sibling-calls. Each compiles the same generated C and every pinned
runtime/*.c, links -lm, and includes the pinned runtime directory. No allocation
counting macro, sanitizer, architecture flag or LTO flag is added. --count is
the parse driver's argument. [build-commands.json](build-commands.json) records
the full argument arrays, [inputs.json](inputs.json) records source/manifest and
binary hashes, and both build logs are retained.

Exact common Callgrind 3.24.0 cache model, unchanged from the preceding report:

```text
--tool=callgrind --cache-sim=yes --branch-sim=yes
--I1=32768,8,64 --D1=32768,8,64 --LL=268435456,1,64
```

L1 instruction misses are simulated, not hardware counters. Both profiling
processes ran concurrently; no profiler time is used for native timing.
The simulator's cache-description and nonfatal brk-segment warnings remain in
stderr. Both complete profiles have every self-event sum equal to its totals
footer. Their summary Ir is 2 above the self/footer total; all other events
agree. The table uses self/footer values. Each of 13 footer events was increased
by one in a real corrupt profile per binary: all 26 were rejected with
self events differ from totals footer. event-mutants.json retains each failure.

After both profiles completed, ordinary execution used wait4 direct-child
ru_utime, one warm invocation per binary, then ten invocations per binary in
alternating order by round. No profiling, builds or tests overlapped timing.
No startup subtraction is applied. user-time.json retains all 20 samples,
user/system/wall times, binary hashes, commands and load averages. All 22
ordinary invocations return 0 with stdout exactly 0 followed by a newline and
empty stderr. This count-mode observation is not full AST parity. Shared-host
noise remains a limit; neither measured instruction nor best user-time result
supports the proposed optimization here.

Reproduce from the parent report's scratch inputs:

```sh
source /workspace/adamic-tools/env.sh
python3 /workspace/scratch/stack-check-scc/tail-calls/run_profiles.py > /workspace/scratch/stack-check-scc/tail-calls/profile-run.log 2>&1
# Wait for both profiles to finish before ordinary timing.
python3 /workspace/scratch/stack-check-scc/tail-calls/user_time.py > /workspace/scratch/stack-check-scc/tail-calls/user-time.log 2>&1
python3 /workspace/scratch/stack-check-scc/tail-calls/reconcile.py > /workspace/scratch/stack-check-scc/tail-calls/reconcile.log 2>&1
```

## Delivery limits

Current origin/main remains c7991b9 and is already an ancestor of this branch.
Only performance documentation and evidence change in this follow-on.
No test gate was rerun for these data-only additions; the preceding merged
uncached native/oracle gate passed (158.602s / 147.103s), including existing
sanitizer lanes. No new sanitizer or LTO result is claimed. Recursion fixtures
and selective runtime callback attributes are conditional on passing the
measurement gate, so they were not pursued. The Workers sieve was not rebuilt
for this stopped parse-only follow-on. Push is only to codex/stack-check-scc.
