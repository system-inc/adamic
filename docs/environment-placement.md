# Closure environment placement

The base is codex/nested-functions b15216d with codex/call-targets 280fc49
merged first as b6e7738. The profile motivation is not an escape proof:
checkTypeRelatedTo and getFlowTypeOfReference account for about 48% of the
sampled Node allocations in docs/stage3-tsc-profile.md on the profile branch.
No whole TypeScript checker compilation or speed claim follows from this unit.

## Placement and ownership

region.go keys one decision by the exact AllocateEnvironment statement.
MakeClosure edges use Function.Environment, including forwarded ancestor slots.
A flow-insensitive alias fixed point tracks closure values; it never clears a
taint after reassignment, so loops, branch joins and aliases cannot hide an escape.
A second fixed point summarizes escaping parameters. Every CallTargets target
must finish with an argument; ClosureTargets.Unknown and missing summaries keep
the heap. Immutable named nested declarations now have authoritative closure
targets too. All existing Call.Function readers in region.go were migrated and
their four guard exemptions removed.

A store in a field, array, Map, Set, global or captured variable, a return,
a throw, or an unknown retaining operation keeps the counted heap. Captures
propagate taint into the capturing closure. All function bodies are inspected,
so a nested helper returning a new closure that forwards an ancestor slot keeps
that ancestor allocation on the heap. Captured parameters and stores into even
local containers are deliberately conservative. The runtime's synchronous
array and collection loops borrow their callback; an unknown callback target
still keeps heap placement. Ordinary closure invocation borrows its carrier,
and passing a closure argument requires every target's parameter proof.

At most 64 fixed slots (2,592 bytes on this amd64 layout) use stack storage.
The emitted fixed C aggregate contains both the header and the ordered cells.
The environment header has an explicit slot pointer, shared by all placements;
its size grows from 24 to 32 bytes on amd64. Cells keep their 40-byte layout and
order. Heap and region allocations still keep header and cells in one allocation.
This avoids flexible-array subobject bounds and gives ASan ordinary frame objects
to instrument. Larger records use an explicit call-owned
region. Both have references zero, so interior-slot retains and releases never
change a count. Closure identity records still use the counted heap, and existing
ABI retain calls remain; this unit does not elide those calls.

Heap, stack and region destruction share adamic_environment_free_children.
The frame registers its environment or region with releaseScopes, before its
closure values and parameters, so those holders end first. Scope cleanup invokes
the destructor once on each actual exit path, including a propagated throw.
Regions walk environment records as well as ordinary objects. Slots retain
ordinary heap references exactly as before; uncounted storage does not make
its contents unowned. There is no early destruction before the call ends.

## Executable evidence

Nineteen environment_*.a oracle fixtures cover direct helpers and local aliases;
forEach, map, filter, sort and find; a helper called in a loop; siblings calling
siblings; returning or storing in a field, array, Map, Set or global; a closure
captured by an escaping closure; keeping named calls and unknown closure calls;
a virtual override that keeps the callback; a callback that creates an escaping
closure; unknown callback targets; normal return, propagated throws and finally;
and a 65-slot call-region layout with both return and throw exits.
TestEnvironmentPlacement asserts the allocation site's placement independently
of output. The oracle executes original source on Node, generated JavaScript
on Node, sanitized native and the release build, then checks leaks and counts.
Existing nested_, closures and regions fixtures are also included in the gate.

Source lowering only throws Error objects and has no parallel-pool or async-frame
IR. Executable three-way fixtures throwing a closure or handing it to those
facilities are therefore not possible on this base. A synthetic IR throw proves
that a thrown carrier stays on the heap. A synthetic unknown transfer proves the
fallback for a future pool or async operation. These are proof tests, not claims
of executable async, pool or arbitrary-throw support. Variable environment
layouts also do not exist in this IR.

## Mutation evidence

| Mutant | Observed catcher |
| --- | --- |
| Replace the returned fixture's heap environment with uncounted stack storage | AddressSanitizer stack-use-after-return reading the captured slot's ready bit. The mutant prevents inlining so this is a return, not a scope, failure. |
| Omit the one reference slot's destruction only on exceptional exits | Node output still agrees and ASan/UBSan pass with leaks disabled; LeakSanitizer reports 144 bytes leaked in two allocations. |
| Read Call.Function in environmentCallEscapes | TestCallTargetReaders rejects the unapproved reader in region.go, exit 1. Production source restored immediately. |

An intermediate __builtin_alloca construction caught a dangling captured heap
string, but did not expose stack-use-after-return to ASan. It was rejected,
without weakening the mutant, in favor of the fixed aggregate above.

The first two are permanent tests in environment_placement_test.go. The direct
reader mutant was run temporarily and restored; the guard has no remaining
region.go exemptions. Logs are /tmp/environment-placement-mutants-fixed-stack.log
and /tmp/environment-placement-guard-mutant.log. An initial mutant run had an
oracle path-normalization error after ASan caught the stack escape; the final
run uses absolute fixture paths and both checks pass.

## Counts and peak memory

Before is a release compiler built from the prerequisite merge, using the same
Go dependencies, clang and Node. After is the placement implementation. Each
entry is before -> after; peak is peak live counted values, not bytes. RSS is
Linux wait4 ru_maxrss in KiB from one separate run of each counted fixture.
Small RSS differences are noisy and do not establish a memory improvement.
Panic fixtures terminate before ordinary cleanup and intentionally do not have
allocations equal frees. The current complete count table is [internal/oracle/counts.md](../internal/oracle/counts.md).

GNU time is absent. A small C fork/exec/wait4 launcher records child RSS and
CLOCK_MONOTONIC elapsed time. Initial Python-only wait4 measurements included
a roughly 9.5 MiB launcher floor and were replaced with the isolated launcher
measurements below. Raw results are under
/workspace/scratch/environment-measurements/{counts,timing}.json and logs are
/tmp/environment-placement-measurements.log. The complete final measurement
run was refreshed after selecting the fixed aggregate storage.

| Fixture | Allocations | Frees | Retains | Releases | Peak live | Regions | RSS KiB |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| environment_array.a | 5 -> 5 | 5 -> 5 | 7 -> 7 | 11 -> 11 | 5 -> 5 | 0 -> 0 | 900 -> 812 |
| environment_callback_escape.a | 7 -> 7 | 7 -> 7 | 5 -> 5 | 11 -> 11 | 6 -> 6 | 0 -> 0 | 896 -> 868 |
| environment_callbacks.a | 14 -> 13 | 14 -> 13 | 12 -> 12 | 14 -> 14 | 13 -> 12 | 0 -> 0 | 972 -> 872 |
| environment_capture.a | 8 -> 8 | 8 -> 8 | 6 -> 6 | 12 -> 12 | 7 -> 7 | 0 -> 0 | 900 -> 788 |
| environment_direct.a | 5 -> 4 | 5 -> 4 | 3 -> 3 | 7 -> 7 | 4 -> 3 | 0 -> 0 | 1,024 -> 784 |
| environment_exits.a | 19 -> 16 | 19 -> 16 | 20 -> 20 | 32 -> 32 | 7 -> 6 | 0 -> 0 | 820 -> 832 |
| environment_field.a | 6 -> 6 | 6 -> 6 | 6 -> 6 | 11 -> 11 | 6 -> 6 | 0 -> 0 | 780 -> 900 |
| environment_global.a | 4 -> 4 | 4 -> 4 | 4 -> 4 | 9 -> 9 | 4 -> 4 | 0 -> 0 | 836 -> 800 |
| environment_keeping_call.a | 4 -> 4 | 4 -> 4 | 4 -> 4 | 9 -> 9 | 4 -> 4 | 0 -> 0 | 968 -> 836 |
| environment_large.a | 395 -> 395 | 395 -> 393 | 133 -> 133 | 397 -> 525 | 131 -> 131 | 0 -> 2 | 1,040 -> 984 |
| environment_local_call.a | 4 -> 3 | 4 -> 3 | 3 -> 3 | 6 -> 6 | 4 -> 3 | 0 -> 0 | 964 -> 896 |
| environment_loop.a | 3 -> 2 | 3 -> 2 | 1 -> 1 | 3 -> 2 | 2 -> 1 | 0 -> 0 | 816 -> 900 |
| environment_map.a | 5 -> 5 | 5 -> 5 | 7 -> 7 | 11 -> 11 | 5 -> 5 | 0 -> 0 | 800 -> 812 |
| environment_returned.a | 4 -> 4 | 4 -> 4 | 4 -> 4 | 8 -> 8 | 4 -> 4 | 0 -> 0 | 932 -> 900 |
| environment_set.a | 6 -> 6 | 6 -> 6 | 7 -> 7 | 11 -> 11 | 5 -> 5 | 0 -> 0 | 820 -> 820 |
| environment_siblings.a | 4 -> 3 | 4 -> 3 | 3 -> 3 | 5 -> 4 | 3 -> 2 | 0 -> 0 | 868 -> 868 |
| environment_unknown_call.a | 5 -> 5 | 5 -> 5 | 5 -> 5 | 11 -> 11 | 5 -> 5 | 0 -> 0 | 916 -> 988 |
| environment_unknown_callback.a | 4 -> 4 | 4 -> 4 | 2 -> 2 | 5 -> 5 | 3 -> 3 | 0 -> 0 | 908 -> 904 |
| environment_virtual_call.a | 5 -> 5 | 4 -> 4 | 5 -> 5 | 10 -> 10 | 5 -> 5 | 1 -> 1 | 1,028 -> 824 |
| nested_array.a | 17 -> 17 | 17 -> 17 | 18 -> 18 | 28 -> 28 | 9 -> 9 | 0 -> 0 | 776 -> 936 |
| nested_captures.a | 22 -> 20 | 22 -> 20 | 8 -> 8 | 22 -> 22 | 7 -> 6 | 0 -> 0 | 968 -> 832 |
| nested_destructured.a | 13 -> 13 | 13 -> 13 | 8 -> 8 | 18 -> 18 | 6 -> 6 | 0 -> 0 | 1,024 -> 940 |
| nested_destructured_tdz.a | 2 -> 1 | 0 -> 0 | 1 -> 1 | 0 -> 0 | 2 -> 1 | 0 -> 0 | 832 -> 772 |
| nested_hoisting.a | 2 -> 2 | 2 -> 2 | 0 -> 0 | 2 -> 2 | 1 -> 1 | 0 -> 0 | 772 -> 872 |
| nested_minimal.a | 2 -> 2 | 2 -> 2 | 0 -> 0 | 2 -> 2 | 1 -> 1 | 0 -> 0 | 896 -> 772 |
| nested_mixed.a | 7 -> 7 | 7 -> 7 | 10 -> 10 | 13 -> 13 | 5 -> 5 | 0 -> 0 | 780 -> 940 |
| nested_mutual.a | 10 -> 8 | 10 -> 8 | 46 -> 46 | 52 -> 50 | 5 -> 4 | 0 -> 0 | 908 -> 852 |
| nested_pattern_parameter.a | 6 -> 6 | 6 -> 6 | 5 -> 5 | 10 -> 10 | 5 -> 5 | 0 -> 0 | 876 -> 904 |
| nested_returned.a | 21 -> 21 | 21 -> 21 | 10 -> 10 | 29 -> 29 | 11 -> 11 | 0 -> 0 | 1,016 -> 1,028 |
| nested_tdz.a | 2 -> 1 | 0 -> 0 | 1 -> 1 | 0 -> 0 | 2 -> 1 | 0 -> 0 | 780 -> 900 |
| nested_tdz_write.a | 2 -> 1 | 0 -> 0 | 1 -> 1 | 0 -> 0 | 2 -> 1 | 0 -> 0 | 820 -> 800 |
| nested_three_levels.a | 8 -> 8 | 8 -> 8 | 7 -> 7 | 13 -> 13 | 7 -> 7 | 0 -> 0 | 988 -> 868 |
| nested_weak.a | 4 -> 4 | 4 -> 4 | 5 -> 5 | 10 -> 10 | 4 -> 4 | 0 -> 0 | 1,024 -> 912 |
| closures.a | 58 -> 58 | 58 -> 58 | 48 -> 48 | 84 -> 84 | 26 -> 26 | 0 -> 0 | 996 -> 960 |
| closures_throw.a | 418 -> 418 | 418 -> 418 | 1,571 -> 1,571 | 1,806 -> 1,806 | 142 -> 142 | 0 -> 0 | 1,016 -> 1,024 |
| closures_throw_uncaught.a | 31 -> 31 | 29 -> 29 | 20 -> 20 | 36 -> 36 | 13 -> 13 | 0 -> 0 | 1,016 -> 940 |

## Million-call benchmark

bench/environment.a has five nested helpers sharing scratch locals and calls
the enclosing function one million times. All native and Node runs exit 0 and
print 52941036. Node runs a temporary .mts copy of the identical .a source with
its built-in type stripping; it does not run Adamic-generated JavaScript.
Release builds use clang -O2; timing builds have neither sanitizers nor counters.
The counts below come from separate --count builds.

| Version | Allocations | Frees | Retains | Releases | Peak live |
| --- | ---: | ---: | ---: | ---: | ---: |
| Native before | 6,000,001 | 6,000,001 | 31,000,000 | 12,000,001 | 6 |
| Native after | 5,000,001 | 5,000,001 | 31,000,000 | 11,000,001 | 5 |

One million counted environment allocations and frees disappear. Five million
closure identity records remain. Retain calls are unchanged, though the calls
on the uncounted record no longer update a count.

Five interleaved rounds ran during the repository gate on this shared four-core
quota. These timings have substantial contention; they are observations, not a
stable speedup claim. The best changed native time is 1.72x Node.

| Version | Best seconds | Median seconds | Range seconds | Peak RSS range KiB |
| --- | ---: | ---: | --- | --- |
| before | 0.383204 | 0.519510 | 0.383204 to 0.722685 | 788 to 856 |
| after | 0.327576 | 0.587279 | 0.327576 to 0.605928 | 668 to 900 |
| node | 0.190800 | 0.300449 | 0.190800 to 0.417117 | 51132 to 51824 |

To repeat the main comparison, build bench/environment.a with the prerequisite
compiler and the changed compiler using `adamic build bench/environment.a -o
<binary>`, then build each again with `--count` for deterministic counters.
Run Node on a temporary .mts copy of the source. Use five interleaved rounds
and a child-RSS meter; never compare the compiler processes' own peak memory.

## Validation

The toolchain setup log reports go ready 0s, clang ready 1s, Node ready 1s,
submodules ready 1s, build cache warm 101s, done 102s. nproc is 5; cpu.max is
400000 100000. Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0. Every shell sources
/workspace/adamic-tools/env.sh. The baseline worktree build uses -buildvcs=false
because its shared cohere submodule link prevents VCS stamping. Test output
always goes directly to a log file.

| Command | Observed result and log |
| --- | --- |
| `bash cloud/setup.sh` | Exit 0; timings above. /tmp/environment-placement-setup.log |
| `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts` | Pass, 11.997s. /tmp/environment-placement-counts.log |
| `go test ./internal/ir ./internal/native -run 'TestCallTarget\|TestClosureTargets\|TestEnvironment\|TestThrownEnvironment\|TestInheritanceMemoryPlans' -count=1` | IR 5.349s, native 3.422s pass. /tmp/environment-placement-targets-final.log |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/ir ./internal/lower -count=1 -timeout 15m` | Final fixed aggregate: native 170.184s, IR 1.855s, lower 29.579s pass. /tmp/environment-placement-packages-fixed-final.log |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestEnvironment\|TestNested\|TestNativeAgreesWithNode/internal/oracle/testdata/(environment_\|nested_\|closures\|regions)\|TestCountsAreRecorded' -count=1 -timeout 15m` | Final fixed aggregate: pass, 44.641s. /tmp/environment-placement-oracle-fixed-final.log |
| `go test ./internal/oracle -run '^TestEnvironment(EscapeStack\|ThrowRelease)Mutant$' -count=1 -v` | Fixed aggregate: pass, 27.295s including cold runtime compilation. /tmp/environment-placement-mutants-fixed-stack.log |
| `go vet ./...` | Exit 0, no diagnostics. /tmp/environment-placement-vet-final.log |
| `gofmt -l cmd internal` | Exit 0, no files. /tmp/environment-placement-format-final.log |
| `git diff --check` | Exit 0, no diagnostics. |

Earlier exploratory fixture runs failed at TypeScript checking or lowering:
TypeScript narrowed directly read saved globals to never, a first-class sibling
reference was NotYet, and capturing a function-body alias in the same record
was correctly refused as a cycle. The final probes read globals through a named
function and use a separate block capture or a newly created closure capturing
scratch. No compiler guard was relaxed to make these fixtures pass.

`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...` was started
on the earlier storage representation. It was manually stopped after 27m45s
while markdownblocks and typeaware suites were still active; there were no
reported test failures. The bridge, all internal packages including the complete
oracle, Unicode properties, CSS, formatfiles, gitignore, GraphQL, JSON and lint
had reported passing results. Log: /tmp/environment-placement-full-gate.log;
stop record: /tmp/environment-placement-full-gate-status.log. This is a partial
development gate, not a full final-representation pass. The required final
package and filtered gates above were rerun for the fixed aggregate. No full
post-change stage1 gate, macOS run, variable-layout run, parallel or async
execution, arbitrary thrown-closure execution, or native tsc compilation is
claimed.
