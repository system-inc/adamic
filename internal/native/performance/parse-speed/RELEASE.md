# Last-reference release boundary, October 7, 2026

Built: a runtime-only release fast path; destruction remains iterative and happens at the same lifetime boundary.
Commits: follows 5ca9162's destruction diagnosis and runtime-area merge; no emission/parser/scanner changes.
Commands and outputs: release control PASS; before 7,389,521,274 Ir, after 6,570,598,637 Ir; 77-file batch8 output byte parity PASS.
Mutant: skipping destroy_last_reference compiles and fails the existing last-release live-value check, exit 1.
Not covered: implemented per-file region, full repository gate, or either dated parse target.

## Exact release build and instruction result

Every Ir number below uses clang 20.1.8, `-O2 -g`, sanitizers off,
`-ffp-contract=off`, `-fno-optimize-sibling-calls`, no LTO, no counted-runtime
hooks. Both measurements use the identical generated runtime-parse.c from
3cbc640; only heap.c changes. Exact after command from the repository root:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/parse-speed/runtime-parse.c internal/native/runtime/*.c -lm -o /workspace/scratch/parse-speed/release-parse > /tmp/parse-speed-release-clang.log 2>&1
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/parse-speed/release.callgrind /workspace/scratch/parse-speed/release-parse --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/parse-speed/release.stdout 2> /workspace/scratch/parse-speed/release.stderr
```

| Same 77 files and release flags | Before | After | Saved |
| --- | ---: | ---: | ---: |
| Whole parse process | 7,389,521,274 | 6,570,598,637 | 818,922,637 |
| Release/child destruction, source-aware v2 self | 1,534,667,356 | 715,744,719 | 818,922,637 |
| Caller cleanup release edges inclusive | 354,105,637 | 354,100,093 | 5,544 |

The reduction is **11.0822%** from the merged runtime. Go parse baseline is
1,684,637,174 Ir (Go 1.27.1 ordinary optimized build, no sanitizer/race,
GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 under Callgrind); native remains **3.9003x**.

## Why

The old release body includes the queue and type-specific destruction switch,
so even releasing an immortal or still-held reference pays its register saves,
draining flag checks/stores and empty-queue loop. The last-reference helper is
noinline so clang cannot put the large body back in the single caller. Ordinary
release now loads the reference count, ignores NULL/immortal values, decrements
a counted value, and enters destruction only at zero. Existing let_go callback
queueing is unchanged. Every last reference is queued and drained before the
outer release returns; reentrant release still queues into an active drain.
There is no recursive free, early free, region, borrowed lifetime or count elision.

The after assembly has no saved-register prologue for ordinary releases. An
immortal release takes five instructions including return; a non-final counted
release takes nine. The stack adjustment/call is reached only at zero. Self
release is 314,959,020 Ir and self destroy_last_reference 182,800,912 Ir, versus
1,316,682,569 Ir self release before. Child field walking and per-file teardown
costs remain essentially unchanged. This confirms the diagnosis: the large
saving is everyday reference bookkeeping, rather than making final-tree teardown
one stroke. Per-file regions remain a separate possible unit with their own
ownership and memory-peak proof.

## Validation

TestRuntimeReleasePaths passes its release (`-O2`, sanitizers off, Count=true)
and sanitizer (`-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`,
Count=true) configurations; common strict C/warning/ffp/sibling flags are those
above. It includes NULL, immortal, shared heap string, final heap string and a
100,000-object chain of heap-built labels. Count semantics are unchanged.

An isolated scratch runtime mutant replaces the new helper call by `(void)value`.
It compiles with the same strict release flags plus `-DADAMIC_COUNT`, then the
existing test's C source fails its live-value check, exit 1:

```
last release left 1 values
adamic: counts: allocations 1 frees 0 retains 1 releases 4 peak 1 regions 0
```

It is caught by the intended destruction check, not a warning, refusal or literal
lifetime accident. Production source is never mutated. Full batch8 release output
is **11,442,907 identical bytes** against its retained Go oracle output on 77
files. An initial comparison mistakenly used Go's four-byte count-mode output;
comparing the full outputs passes and hashes match. Whole AST parity and complete
uncached native/oracle package checks are recorded in the accompanying logs,
including ASan/UBSan/LeakSanitizer. Full repository gate is not claimed.
