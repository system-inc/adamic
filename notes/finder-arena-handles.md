Built a cycle-reach trace, removed the spurious bottom-array edge, and added a numeric-handle arena oracle fixture.
Implementation: b0012cb193bfd23b23d11314ad078a9acb34e0c7; current-main merge: 13c5dc5ac4e68039acb9dec5038db68a2d4e4f0b.
Commands: exact wave11 source lowers to C; uncached lower/oracle packages, focused arena oracle, vet and formatting checks pass.
Mutants: parent-array, hidden-owner and hidden-closure programs stay refused; four independent compiler mutants fail their intended assertions.
Not covered: the full repository gate, the original wave11 input corpus at runtime, or changes to graph-region ownership.

The unit's explicit base was origin/area/compiler at
5ecb2b680c2e9e43f6a76fdc14217afa8102e3b5, rather than the generic main-base
instruction. Before verification, origin/main at
71d7e491b3c9724f7a0e2ee754592149e7f9790b was merged into this branch.
No protected compiler file was edited, and no cohere code was copied.

The observed original path, printed by `ADAMIC_TRACE_CYCLES=1` from the
finder's actual breadth-first search, was:

```text
cycle reach: CollapseCopyNode[] -- slot contents --> CollapseCopyNode
  CollapseCopyNode -- field children --> CollapseNodeList
  CollapseNodeList -- assignable program shape --> { nil: true; length: number; capacity: number; values: never[]; }
  { nil: true; length: number; capacity: number; values: never[]; } -- field values --> never[]
  never[] -- related holder types --> CollapseCopyNode[]
```

The source was the exact Adamic repository program at
`98f7e5f08c10f148d457e3a1fdde0355d1401f6c`,
`stage1/cohere/lint/helpers/gaps/slot01_wave11_main.a`. Its helpers were extracted
from that repository branch into `/tmp/finder-wave11` using `git archive`.
The refusal pointed to `main.a:51:11` and named the push at `main.a:58:9`.
The original trace and diagnostic are preserved in
`/tmp/finder-arena-original.log`.

The last edge is spurious. The checker allows `never[]` to be assigned to
`CollapseCopyNode[]`. That relation does not establish an owning path: the
empty backing array cannot contain any element, and the concrete list shape
has no arena or function field. The structural shape edge preceding it is
legitimate; the finder then incorrectly identifies the backing array with an
unrelated arena holder. No function type or closure capture occurs on this
path, so the allocate closure is not the cause.

The fix stops reach at arrays whose element type is `never`, before the
symmetric holder-type relation is tested. Primitive handles already stop at
the existing non-object check. The exception relies on the existing invariant
mutable-view gate: a retained `never[]` cannot be populated through a wider
mutable alias in an accepted program. The new lowering test explicitly checks
that `const bottom: never[] = []; const values: number[] = bottom;` is refused
by `adamic/invariant-mutable`. Intersections still traverse their members,
and normal object shapes, fields, collections, cells and closures remain
followed. This is a narrow repair to the observed edge, not a claim that every
possible structural false positive has been eliminated.

Tracing is opt-in and writes successful reach paths to stderr, naming each
field, structural match, collection argument and captured cell. Closure edges
also name their source location. Parent-path storage is allocated only when
tracing is enabled. Normal diagnostics are unchanged.

The authored fixture `internal/oracle/testdata/numeric_handle_arena.a` has the
same strings, booleans, string map and numeric child-list record shape, a
contextually typed empty list, and an allocate closure that captures its arena.
It is registered in a separate oracle test file, following the existing
fixture-registration pattern. Node, the JavaScript backend, release native,
ASan/UBSan native and LeakSanitizer agree. Source Node prints:

```text
0:0
1:1
2:1
n1:entry1
3:changed:true
```

The exact wave11 source also lowers successfully after the fix and main merge:
`go run ./cmd/adamic c /tmp/finder-wave11/stage1/cohere/lint/helpers/gaps/slot01_wave11_main.a`.
Its generated C is `/tmp/finder-arena-fixed.c` (179516 bytes), with empty
stderr in `/tmp/finder-arena-final-source.log`. This observation establishes
that the reported refusal is removed; it does not establish runtime agreement
for the entire wave11 input corpus.

The parent mutant adds `parent: CollapseCopyNode[]` and initializes it with
`arena`. The real closing path is:

```text
CollapseCopyNode[] -- slot contents --> CollapseCopyNode
CollapseCopyNode -- field parent --> CollapseCopyNode[]
CollapseCopyNode[] -- related holder types --> CollapseCopyNode[]
```

Its entire diagnostic is pinned in `TestNumericHandleArena`, normalizing only
the test's temporary directory:

```text
main.a:10:5: Adamic 0.1 refuses CollapseCopyNode[], an array whose elements can reach back to an array like it: a cycle reference counting can't free, and the write at main.a:27:9 may close one (the value written reaches something this function didn't make or let escape, and what it's written into wasn't made here); declare the elements weak, Weak<CollapseCopyNode>[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly CollapseCopyNode[]; or write into such an array only values this function made, or only into one it made (adamic/cycle-capable)
```

Two further program mutants give a structurally compatible list an extra
`owner: arena` or `callback: () => arena.length`, using a `number[]` backing
so invariant widening cannot mask the cycle check. Both are refused by
`adamic/cycle-capable`. Their actual paths remain printed in
`/tmp/finder-arena-traces.log`; the callback path follows its field, the
closure capture, `cell arena`, and the cell's array contents.

Four independent compiler mutants were run with
`go test ./internal/lower -run TestNumericHandleArena -count=1 -v`.
Each was restored before proceeding; each exited 1 for its intended assertion,
with no build-error kill:

| Compiler mutant | What caught it | Log |
|---|---|---|
| Remove the new never-array stop | Valid arena falsely refused | `/tmp/finder-arena-mutant-bottom.log` |
| Disable related holder matching in reach | Parent-array mutant accepted | `/tmp/finder-arena-mutant-real-path.log` |
| Disable assignable program-shape edges | Hidden owner and hidden callback accepted | `/tmp/finder-arena-mutant-structural.log` |
| Disable closure matching | Hidden callback accepted | `/tmp/finder-arena-mutant-closure.log` |

Setup used `export GOPROXY='https://proxy.golang.org|direct'` before
`bash cloud/setup.sh`, with output in `/tmp/finder-arena-setup.log`.
It succeeded and printed `/workspace/adamic-tools/env.sh`, sourced in every
build and test shell. `nproc` was 5, with cgroup `cpu.max` 400000/100000.
The timing lines were:

```text
setup: submodules ready (0.101s)
setup: go ready (0.393s)
setup: node ready (0.419s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1.028s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=1.036s
setup: markdown dependencies ready (1.668s)
setup: go build ready (64.476s)
setup: test binaries deferred (use --warm-tests) (64.608s)
setup: build cache warm (64.611s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (64.735s)
```

Go was 1.27.1, Node 24.19.0, clang 20.1.8. The initial fetch only fetched main;
explicitly fetching the named compiler and helper branches resolved the
missing remote-tracking refs.

All test output went directly to log files. Validation commands and results:

```text
go test ./internal/lower -run TestNumericHandleArena -count=1 -v
ok github.com/system-inc/adamic/internal/lower 0.216s
ADAMIC_TRACE_CYCLES=1 go test ./internal/lower -run TestNumericHandleArena -count=1 -v
passed, with parent, structural-owner and closure paths printed
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/numeric_handle_arena' -count=1 -v -timeout 30m
ok github.com/system-inc/adamic/internal/oracle 0.658s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 27.311s
go vet ./...
exit 0, no diagnostics
gofmt -l cmd internal
exit 0, no output
git diff --check
exit 0, no output
```

Logs are `/tmp/finder-arena-focus.log`, `/tmp/finder-arena-traces.log`,
`/tmp/finder-arena-oracle.log`, `/tmp/finder-arena-counts.log`,
`/tmp/finder-arena-vet.log`, and `/tmp/finder-arena-gofmt.log`.
The counts update adds only the new fixture row: 51 allocations, 51 frees,
38 retains, 51 releases, peak 30, and 0 values in regions. No existing row moved.

The complete touched packages passed after merging current main, with all
compiler mutants restored:

```text
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/oracle
ok github.com/system-inc/adamic/internal/lower 54.716s
ok github.com/system-inc/adamic/internal/oracle 224.363s
```

The full output is `/tmp/finder-arena-packages.log`. This includes the existing
cycle and freshness refusal probes and all oracle fixtures, with source Node,
backend Node, release native, ASan, UBSan, leak checks and recorded counts.
The full repository test gate was not run under the worker-gate exception.
