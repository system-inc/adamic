# GCC runtime warning audit

Base: `044899ef4ba57ed211075eb525e5ef1872bb7f0d` (`runtime/area-take-leftovers`).
Compiler: Debian GCC 14.2.0-19, flags `-std=c11 -O2 -Wall -Wextra`.

## Numeric invariants

`dtoa.c:153`: GCC inferred a possibly negative starting index in the excess-digit
clearing loop. The implementation follows V8's
[AssignBignum and Clamp](https://github.com/v8/v8/blob/13.6.233.17/src/base/numbers/bignum.cc).
Digit counts start at zero, assignments and growth produce nonnegative counts,
and Clamp decrements only while the count is positive. A negative count is
unreachable. Guard `other->used_digits < 0` at AssignBignum's entry so GCC can
prove the clearing loop's lower bound without following every bignum operation.

`ieee754.c:632`: GCC could not prove that `fq[0]` is initialized. In
[fdlibm k_rem_pio2.c](https://github.com/freemint/fdlibm/blob/master/k_rem_pio2.c),
when z is zero, the recomputation check requires at least one nonzero iq term
at or above jk before proceeding. Trimming zero terms therefore cannot make
jz negative. If z is nonzero, jz stays at least its positive initial jk.
Guard `jz < 0` before the conversion/product loops, making the existence of
`fq[0]` explicit to GCC. Neither guard changes a valid number's output.

## Map iterator count

The original diagnostic was:

```text
map.c:248: warning: '__atomic_fetch_add_8' writing 8 bytes into a region of size 0 overflows the destination [-Wstringop-overflow=]
cc1: note: destination object is likely at address zero
```

The original code was:

```c
adamic_map_iterator *adamic_map_iterate(adamic_map *map) {
    adamic_map_iterator *iterator = adamic_allocate(sizeof *iterator, adamic_kind_map_iterator);
    iterator->map = adamic_retain(map);
    iterator->next = 0;
    iterator->exhausted = false;
    map->iterating++;
    if (adamic_graph_is(map)) { iterator = adamic_graph_adopt_owned(iterator, sizeof *iterator); }
    return iterator;
}
```

This write is not a race: `iterating` is `_Atomic size_t` in adamic.h,
and the increment, reads, exhaustion decrement and destruction decrement all
use C11 atomic operations. The compiler permits cross-worker Map captures only
when readonly (docs/concurrency.md); worker-local iterators share this atomic
tally, not their mutable cursor. The warning concerns a potential null map:
adamic_retain accepts NULL as undefined, whereas map iteration requires a map.
A local `map == NULL` guard makes that precondition explicit and removes the
warning. The atomic field and its memory ordering are unchanged.

## Other diagnostics

GCC 14.2 reports no `-Winfinite-recursion` warning when compiling all 56 runtime
C translation units. No recursion change is warranted.

GCC ignores `#pragma STDC FP_CONTRACT OFF` in hypot.c, ieee754.c and radix.c.
The pragma is retained for compilers that honor it; GCC requires the build flag.
The existing `TestRuntimeCompilesEveryUnitUnfused` audits all runtime compile
commands in release, sanitized oracle and count builds. It also removes
`-ffp-contract=off` from dtoa.c's actual compile command and requires the audit
to reject that omission. This coverage was already present on the base branch.

## Validation and measurement

Machine: Linux x86_64 KVM, Intel Xeon Platinum 8573C, 5 visible vCPUs with a 4-CPU cgroup quota
(`cpu.max=400000 100000`); Node
v24.19.0. Set up with `GOPROXY=https://proxy.golang.org|direct`,
`bash cloud/setup.sh --wasi-sdk`, and `/workspace/adamic-tools/env.sh`.

A Python monotonic timer around a sequential subprocess compile of each of the
56 runtime C files measured 23.320 seconds after the changes. Only the three
ignored-pragmas diagnostics remain: no array-bounds, maybe-uninitialized,
stringop-overflow or infinite-recursion warnings. Load averages at the end were
17.27 / 5.84 / 2.11. Setup and targeted Go builds overlapped this run; this is
compiler validation timing, not a performance comparison.

No fixtures or mutants were added or changed; counts.md regeneration is not
required. Targeted test results are recorded below.

Passed commands:

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(number|math|dtoa|parse)' -v
go test ./internal/native -run '^(TestRuntimeCompilesEveryUnitUnfused|TestLibraryMapSetIteratorResources)$' -v
go test ./internal/native -run '^TestParallelMap$/^tsan$/^map$' -v
```

The oracle matched Node on all 21 selected fixtures (92.239 s test process).
The native checks passed (94.693 s): all 56 compile lines were protected in each
of the three modes, and the dtoa-only omission mutant was rejected. The map
resource check reported 788 KiB peak for both control and held exhausted
iterators, and 0.036411 CPU seconds for two million numeric lookups.
The existing shared-map iteration harness passed under TSan, with one and four
workers, three runs each, no race reports (27.692 s test process).

Bash TIMEFORMAT measured whole-command elapsed times of 497.076 s (oracle),
497.021 s (native audit/resources) and 374.384 s (TSan). These include cold Go
compilation and overlapping setup/build work, not just test execution. Oracle
load averages were 6.12 / 2.03 / 0.74 before and 10.37 / 14.46 / 8.59 after.
Setup completed successfully in 780.704 s; no whole-package test run was used.
