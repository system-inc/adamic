# Reliable concurrency race checks

This unit starts at concurrency-scaling 0f67284. Every final runtime negative
control was caught in 50 of 50 checks. Each of the five TSan controls was caught
in all 150 individual executions. A race control must produce a TSan data-race
report, not merely crash or time out.

Claimed files: internal/native/parallel_test.go, internal/native/scaling_test.go,
new internal/native race-check test helpers, internal/native/testdata/parallel/*.c,
internal/native/native.go (TSan-only flag), internal/native/runtime/adamic.h,
internal/native/runtime/heap.c, internal/native/runtime/parallel.c,
internal/native/runtime/object.c (cache publication hook), new runtime
TSan hook files, and internal/oracle/parallel_test.go. No lowering or JavaScript
changes. Release and ASan builds must contain no hook code; release objects are
compared byte for byte with the parent.

At this parent there are five TSan race controls, six C assertion controls and
two scaling guard/leak controls. All thirteen are measured; their detectors are
reported separately rather than calling assertion failures race reports.

## Rates

The first, lighter-load baseline caught all thirteen controls 50/50. Repeating
that unchanged parent in an isolated worktree alongside the gate reproduced the
escape: missing items sharing was caught 48/50. Both misses exited 0 and printed
`strings 2048 5242880`. The other parent controls stayed 50/50.

| Control | Detector | Parent under gate load | Final checks | Final TSan executions |
|---|---|---:|---:|---:|
| skip_items_share | TSan data race | 48/50 | 50/50 | 150/150 |
| plain_shared_count | TSan data race | 50/50 | 50/50 | 150/150 |
| field_cache | TSan data race | 50/50 | 50/50 | 150/150 |
| remote_free | TSan data race | 50/50 | 50/50 | 150/150 |
| result_order | C comparison, exit 3 | 50/50 | 50/50 | n/a |
| reused_graph | C lifecycle assertion | 50/50 | 50/50 | n/a |
| one_worker | C assertion, abort | 50/50 | 50/50 | n/a |
| lazy_cache | TSan data race | 50/50 | 50/50 | 150/150 |
| oversized_slot | C assertion | 50/50 | 50/50 | n/a |
| fixed_grain | C assertion | 50/50 | 50/50 | n/a |
| eager_strings | C assertion | 50/50 | 50/50 | n/a |
| pointer_guard | expected panic status/message | 50/50 | 50/50 | n/a |
| loser_free | LeakSanitizer | 50/50 | 50/50 | n/a |

The first seven rows are the original runtime controls. Four are race checks;
the other three use their original comparison or assertion. Lazy cache publication
adds the fifth TSan control. Calling all seven TSan races would misstate what was
run. The additional scaling controls were measured too.

An intermediate cache fixture also escaped: 42/50 three-execution checks passed,
with 132 catches in 140 attempted executions. A miss stops that check immediately;
later attempts are not silently substituted. After isolating field reads from
method-cache operations and pausing after cache publication, a separate 50-check
cache run caught 150/150, followed by the full final run above.

These are observations. The inference behind separating the cache phase is that
other cache operations can introduce ordering in TSan's model and mask the field
accesses. No claim rests on reproducing that inference on another toolchain.

## Changes holding each race

- Missing sharing: bounded pauses after the unshared count read, at scope
  publication and after a range claim. The claim pause is outside the scheduler
  mutex. The existing first-callback rendezvous remains, before any slice result
  can accidentally share its input owner.
- Plain shared count: a count-only phase before strings, iterators and Weak's
  mutex, plus a pause between the shared count read and increment.
- Remote free: the owner and four releasers rendezvous immediately before the
  allocator phase. All Weak work is finished then. A pause precedes remote-node
  publication, and is still present when the mutant substitutes an owner-list write.
- Field cache: a field-only phase before method-cache operations, and a pause
  after the atomic cache store. The mutant still removes only the atomic read.
- Lazy cache: the existing four-builder gate is retained, forcing competing
  publications and three losing copies. Its mutant is unchanged.

Regular positive TSan harnesses and the TSan variant of every parallel oracle
fixture run three fresh processes. Every run must remain sanitizer clean and byte-identical. Each
race mutant runs three fresh processes too, and **every one must report a data
race**. There is no retry-until-pass rule. Assertion controls retain their original
single execution and criterion.

`TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1` is fixed by both
harnesses. `ADAMIC_THREADS=4` fixes the C pool size, except the one-worker control.
The pool scans victims deterministically and has no random seed to set. An OS
scheduler seed is not available.

`ADAMIC_TSAN_PERTURB=1` enables the optional hook. It sleeps 100 microseconds on the
first 64 visits to each point per thread, preserving errno. It adds no locks,
atomics, or happens-before edges. Budgets keep a million-number map from sleeping
per element. No pool or ownership ABI changed.

## Release and sanitizer exclusion

Only `Options{ThreadSanitize: true}` defines `ADAMIC_TSAN_TEST`. The header refuses
that definition without the actual thread sanitizer. Without it, pause sites
preprocess to nothing observable and the function, TLS budgets and environment
reads are absent.

`check_objects.py` compiled all 47 parent and current release translation units
with the same filenames, paths and release flags: **all 47 objects were byte
identical**. All 47 current ASan/UBSan objects were also compiled and inspected:
no hook definitions or references. Release had none either. ASan debug-object
identity is not claimed because source line tables changed.

The identity check was itself tested with a scratch fixed-grain mutant. Only
`parallel.c` differed; the script exited 1 with `AssertionError: release objects
changed`. Defining the test macro in a release build was rejected with
`ADAMIC_TSAN_TEST requires ThreadSanitizer`.

## Harness wall time

Best of five, with medians in parentheses, in seconds. Each cell times all three
fresh TSan executions, excluding compilation. Linux amd64, AMD EPYC 9V74,
`nproc=5`, cgroup `cpu.max=400000 100000`, 17.6 GB; concurrent native/oracle gates
and mutant runs, one-minute load about 9 to 12. These are gate costs under load.

| Harness | ADAMIC_THREADS=1 | ADAMIC_THREADS=4 |
|---|---:|---:|
| scaling | 0.309 (0.364) | 0.298 (0.357) |
| cache publication | n/a | 2.650 (3.498) |
| slot pairs | n/a | 0.396 (0.465) |
| numbers | 0.077 (0.133) | 0.129 (0.182) |
| lifecycle | 0.764 (0.888) | 0.708 (0.857) |
| strings | 1.098 (1.414) | 0.805 (0.913) |
| objects | 0.253 (0.286) | 0.334 (0.367) |
| map | 0.145 (0.184) | 0.206 (0.256) |
| fresh | 0.131 (0.202) | 0.253 (0.332) |
| nested | 0.190 (0.277) | 0.353 (0.475) |
| exception | 0.186 (0.200) | 0.308 (0.392) |
| nested_exception | 0.085 (0.102) | 0.166 (0.231) |
| million | 0.729 (0.972) | 0.775 (0.866) |
| memory | n/a | 7.233 (8.555) |

The five-round positive harness measurement passed in 79.704 seconds. The complete
final fifty-run control measurement passed in 654.329 seconds, including snapshot
builds and all thirteen controls. Raw check and execution times, rates, representative
TSan summaries and every object hash are in
[race-checks-reliable-results.json](race-checks-reliable-results.json).

## Commands and outputs

All test output went directly to files. Setup printed:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (53s)
setup: done in 53s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. Source the setup environment before each
command: `source /workspace/adamic-tools/env.sh`.

```
# Original checkout, before changing the fixtures. Exit 0, 205.832s.
go test -count=50 -timeout 30m -v \
  -run '^TestParallel(ChecksCatchMutants|ScalingGuardMutants)$' ./internal/native \
  > /workspace/race-reliable-baseline.log 2>&1

# Same command in /workspace/adamic-race-before, detached at 0f67284.
# Exit 1, 225.733s: missing sharing escaped twice; other controls passed.
# Output: /workspace/race-reliable-baseline-loaded.log

# Final implementation. Exit 0, PASS, 654.329s.
go test -count=50 -timeout 30m -v \
  -run '^TestParallel(ChecksCatchMutants|ScalingGuardMutants)$' ./internal/native \
  > /workspace/race-reliable-final-rates.log 2>&1

# Cache fix in isolation. Exit 0, PASS, 57.607s; 150/150 TSan reports.
go test -count=50 -timeout 30m -v \
  -run '^TestParallelChecksCatchMutants$/^field_cache$' ./internal/native \
  > /workspace/race-reliable-cache-window.log 2>&1

# Exit 0, ok github.com/system-inc/adamic/internal/native 231.176s.
go test -count=1 -timeout 30m ./internal/native/... \
  > /workspace/race-reliable-native-final.log 2>&1

# Exit 0, ok github.com/system-inc/adamic/internal/oracle 343.959s.
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle \
  > /workspace/race-reliable-oracle-final.log 2>&1

# Exit 0, PASS, 79.704s; five checks of each positive TSan harness.
go test -count=5 -timeout 30m -v \
  -run '^TestParallel(Map|Memory|Lifecycle|Scaling|CachePublication|SlotPairs)$/^tsan$' \
  ./internal/native > /workspace/race-reliable-harness-times.log 2>&1

# Exit 0; 47/47 identical release objects; no release or ASan hook symbols.
python3 internal/native/testdata/parallel/check_objects.py 0f67284 \
  > /workspace/race-reliable-objects-final.json \
  2> /workspace/race-reliable-objects-final.log

# Expected exit 1: parallel.c object changed, proving the comparison can fail.
python3 internal/native/testdata/parallel/check_objects.py 0f67284 --mutant \
  > /workspace/race-reliable-object-mutant.json \
  2> /workspace/race-reliable-object-mutant.log

# Exit 0; verifies all final rates before saving the artifact.
python3 internal/native/testdata/parallel/summarize_races.py /workspace \
  > docs/race-checks-reliable-results.json

go vet ./... > /workspace/race-reliable-vet.log 2>&1
# Exit 0, no diagnostics.
gofmt -l cmd internal > /workspace/race-reliable-format.log
# No files.
```

The intermediate fifty-run measurement is retained in
`/workspace/race-reliable-after.log`: exit 1, 597.746 seconds, eight cache-check
misses. It is not a final green gate.

Commits: `debb1ae` claims; `dbb48b3` implementation. The report commit follows
these on `codex/race-checks-reliable`, pushed without a PR. This unit changes no
Adamic source files and no lowerer or JavaScript code. macOS arm64 and the async
branch 3a2dbd6 were not run here. The rates describe this Linux toolchain and these
fixtures; the suite now tests each race three times and rejects any single miss.
