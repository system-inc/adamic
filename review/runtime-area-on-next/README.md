Review: six passing .a probes; no production changes.
Branch: runtime/review-area-on-next, based on origin/runtime/area-on-next 48ea5354267465c02416d7a023c0a25f425ecc07.
Validation: source Node, ASan/UBSan, LeakSanitizer and balanced counts; TSan at 4 and 16 threads.
Proofs: eight independent mutants caught; compiler witnesses below remain outside the passing .a files.
Limits: Linux review only; no full repository gate, full oracle, macOS, WASI or Program-region integration.

Read merge messages ab827d7c and 48ea5354 before writing probes. Linux, Go 1.27.1,
clang 20.1.8, Node 24.19.0; nproc 5, cgroup quota 4 CPUs. Setup's cumulative timing
lines: Go 0.026s, Node 0.029s, dependencies 0.085s, submodules 0.114s, clang 0.197s,
build/cache 51.913s, done 51.939s. Full evidence is beside this report. Saved logs are gzip-compressed without timestamps,
preserving the exact source observations and sanitizer diagnostic bytes.

## Passing programs

| Program | Observation | Allocations / frees / retains / releases / peak / in regions |
|---|---|---|
| async_identity.a | f === f inside async, saved identity across await, escaped closure called after return; four true lines, then 9 | 10 / 10 / 24 / 38 / 9 / 0 |
| environment_identity.a | Ordinary environment canonical cache, sibling returning f, escaped closure; three true lines, then 4 | 5 / 5 / 11 / 11 / 4 / 0 |
| inherited_statics.a | Parent, child and leaf observe inherited fields until own writes shadow them; checked scalar field contracts enable readiness/type bytes | 27 / 27 / 25 / 52 / 10 / 0 |
| parallel_shapes.a | 64 tasks, two actual layouts with value at slots 0 and 3; one layout also has an unrelated uninitialized readonly slot. Checked field reads return 2359296 | 10 / 10 / 197 / 138 / 8 / 0 |
| region_partial.a | Constructor adopts a cyclic-capable object before text/count are assigned. Both definite-assignment fields become ready, then reads and a later write agree with Node | 9 / 9 / 12 / 19 / 4 / 0 |
| awaited_void.a | Awaited Promise<void> assigned to a local; prints done then undefined | 10 / 10 / 22 / 37 / 10 / 0 |

Every program exits 0 with empty semantic stderr. The counted lane removes only
its exact counts line and the documented graph-region diagnostic before comparing
stderr. region_partial reports one reachable 80-byte graph value and 16 metadata
bytes, zero unreachable values. Its generated C uses adamic_object_size for adoption,
marks text/count unready, then writes their type and readiness bytes before checked
reads. This covers construction and subsequent graph lifetime, rather than an
arbitrary later relocation of an already populated live object.

The parallel probe runs sanitized, leak-checked and counted at ADAMIC_THREADS=4
and 16. Each TSan setting runs three fresh processes, all matching Node. The
readiness fault can observe Owner.unused's false readiness bit; the representation
fault can observe a neighboring boolean/string instead of value's number byte.

Builds use native.Build with its warning-strict C11 policy, -pthread,
-ffp-contract=off and -fno-optimize-sibling-calls. Sanitized: -O1 -g
-fsanitize=address,undefined -fno-sanitize-recover=all. Counted: -O2 -DADAMIC_COUNT.
TSan: -O1 -g -fsanitize=thread -DADAMIC_TSAN_TEST, with
TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1 and
ADAMIC_TSAN_PERTURB=1. These are review lanes, not ThinLTO shipped builds.
Leak detection runs separately through internal/leakcheck.Check.

## Source witnesses that cannot be pushed as passing programs

A local nested function self-comparison passes checking/lowering but fails both
sanitized and counted C builds with -Werror,-Wtautological-compare. Node prints
true and exits 0. The generated comparison is the same C variable on both sides.
This is a compile failure, not observed wrong native output:

```a
function run(): void {
 function f(): number { return 1; }
 console.log(String(f === f));
}
run();
```

The canonical-closure/graph-region stop is reachable from source. Node prints true
and exits 0; lowering succeeds, then native.C panics with
`native: a canonical closure that joins a graph region was not refused in lowering`.
No native binary exists to sanitize or count:

```a
function make(): () => boolean {
 let saved: (() => boolean) | undefined = undefined;
 function f(): boolean { return saved === f; }
 saved = f;
 return f;
}
console.log(String(make()()));
```

The requested void refusal holds. Node prints done; Adamic refuses line 2 with
`an unawaited async task through the void operator; await the call or keep and return its Promise`:

```a
async function done(): Promise<void> { console.log('done'); }
void done();
```

Initial draft probes also stopped on TypeScript's narrowing of a global assigned
only inside an async function, mutable parallel input ownership, constructor this
escape before initialization, and the existing refusal of `undefined!` initializer
syntax. Those were adjusted to valid probes: read the escaped global through a
function, use a readonly local array, and use definite-assignment property declarations.
They are not reported as sanitizer findings. All final passing files are tested
as committed; no warning policy or production guard was weakened.

## Checks proved by faults

All runtime faults are built from scratch copies, and C faults change only generated
scratch C. The void guard fault uses a Go overlay. None edits production sources.
Every listed mutant compiles; build warnings do not count as catches.

| Mutant | Catcher |
|---|---|
| Skip canonical cache lookup | Native still exits 0 with empty stderr and passes leakcheck; Node comparison catches false across await instead of true |
| Treat async frame as an environment when locating its cache | ASan heap-buffer-overflow in closure.c:88 |
| Index an inherited field's metadata from the child constructor and its slots | UBSan unsigned pointer-offset overflow in object.c:205 |
| Index readiness from the shared packed slot cache | Wrong-field readiness refusal, exit 70, under TSan at both 4 and 16 threads |
| Index representation from the shared packed slot cache | Wrong-field type refusal, exit 70, under TSan at both 4 and 16 threads |
| Omit two metadata bytes per slot from graph adoption's size | ASan heap-buffer-overflow when writing the adopted header's field metadata |
| Omit the escaped closure's final release | Node output still matches; LeakSanitizer finds 440 bytes in two allocations, counted rule reports 10 allocations, 8 frees, 0 in regions |
| Remove the void-of-Promise refusal | Overlay lowering succeeds; refusal checker fails with FAIL void refusal: <nil> |

The cache-index mutants fail semantic checks under TSan, rather than emitting a
TSan data-race report: their cache reads remain atomic, but the returned slot and
later cache contents can belong to different layouts. An initial weaker parent
mutant changed only the metadata buffer while preserving the correct parent slot
index; it survived because inherited layouts preserve those type bytes. The final
mutant reintroduces the actual wrong-owner index and is caught by UBSan.

## Reproduce

From the repository root, with the setup environment sourced:

```sh
source /workspace/adamic-tools/env.sh
go run ./review/runtime-area-on-next/check /tmp/area-review-controls review/runtime-area-on-next/*.a > /tmp/area-review-controls.log 2>&1
go run ./review/runtime-area-on-next/check /tmp/area-review-mutants --mutants > /tmp/area-review-mutants.log 2>&1
python3 review/runtime-area-on-next/check/prove-refusal.py /tmp/area-review-void > /tmp/area-review-void.log 2>&1
go vet ./review/runtime-area-on-next/check > /tmp/area-review-vet.log 2>&1
```

Observed exits: 0 for all four commands. Controls: programs=6,
failed-or-refused=0. Runtime/C proofs: missed=0. Void proof: control exits 0,
mutant exits 1 for the expected accepted-task diagnostic. Vet emits nothing.
The helper executes source Node directly, without oracle observation caches.
