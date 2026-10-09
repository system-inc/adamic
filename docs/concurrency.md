# Concurrency, part 1

Status: **design decided by @system_adamic, approved by Kirk, October 6, 2026** (#p286ycm). This is the contract for the implementation, not a claim that it has landed.

Adamic uses every core by proving what may cross between them. Ordinary code has no raw threads, detached tasks, locks or shared mutable globals. Parallel work is structured fork-join:

```ts
import { parallelMap } from 'adamic';

parallelMap<T, R>(items: readonly T[], work: (item: T, index: number) => R): R[];
```

Every child finishes before the call returns. Results always follow input order. On Node, `parallelMap` is an ordinary sequential map. That is the sequential witness: native stdout, stderr and exit status must match it byte for byte.

## What may cross

The compiler infers Shareable structurally. Numbers, booleans and strings qualify. Records, tuples, arrays, Maps and Sets qualify only when the whole reachable type is readonly and every part is Shareable. Functions qualify only when their captures do. A readonly field holding a mutable Map is still mutable data. There is no unchecked escape hatch.

A task's closure may capture only Shareable values from immutable bindings. Capturing a `let` is refused even when the closure only reads it. A mutable field is refused too. The closure and everything it calls must be proven not to write state another task can reach. Writes to its own locals and fresh objects it allocated are fine; writes to globals, captures or items are refused. Part 1 may conservatively refuse a call whose effects it cannot prove.

Refusals use the existing lowering diagnostics and name the path that broke the proof, for example:

```
task capture 'options' is not shareable: options.cache is a mutable Map
```

Moving mutable values into tasks belongs to part 2. Async and await belong to part 3. Part 1 refuses both clearly rather than guessing their meaning.

## What the runtime owes

Native work runs on a fixed work-stealing pool sized to the available cores. `ADAMIC_THREADS` overrides that size; `ADAMIC_THREADS=1` is supported and is the sequential baseline. The first parallel call starts the pool lazily. Exit joins every worker and leaves the leak check clean. Nested calls must finish without starving their children.

Values start thread-local with plain reference counts. Immediately before items or closure captures become reachable from another worker, the runtime marks their entire reachable graph shared. A header bit records this permanently. Truly immortal constants need no marking. Shared objects use atomic retain and release; unshared objects retain the plain fast path. Region storage must remain alive through the join and must not be mistaken for a static constant merely because its count is zero.

Part 1 marks each result's whole reachable graph shared before publishing it to the parent. This also covers results that alias an item or capture; it needs no unique-ownership assumption. Shared marking must account for string owners and UTF-16 caches, and all runtime mutation hidden behind an immutable source value must be safe. Reuse must never mutate an object another worker can reach. `ADAMIC_COUNT` remains exact across workers.

## What holds it honest

The oracle runs parallel fixtures against Node, under ASan and UBSan with a leak check, and separately under ThreadSanitizer on Linux. TSan and ASan are separate builds. Fixtures cover numbers, string views, readonly records and nested arrays, readonly Map captures, fresh object results, ordered merging, one worker and many workers. Each refusal rule has a fixture holding its exact message.

Every check must be shown able to fail: skipping shared marking must produce a TSan race; plain counting on a shared object must fail TSan or the leak check; accepting a mutable capture must fail a refusal fixture; reordered results must fail the Node comparison; losing the one-worker path must fail its fixture.

The real workload generates a few thousand in-memory file strings deterministically, tokenizes and summarizes each through pure functions, then merges results in order. Report wall time at one, two and every available core, best of five, with load stated. Correctness comes before a speed claim.

## Runtime file claims, part 1

The runtime unit owns these changes on `codex/concurrency`: `internal/native/runtime/adamic.h`, `heap.c`, new `heap_parallel.h`, new `parallel.c` and `parallel.h`, new `share.c`, `count.c`, `count.h`, `string_index.c`, `string_append.c`, `exceptions.c`, `stack.c`, `weak.c`, `adamic.c`, `map.c`, `library_language.c`, and `sort.c`; the field-cache declaration hook in `internal/native/emit_objects.go`; sanitizer and pthread build hooks in `internal/native/native.go` and `runtime_library.go`; and new `internal/native/parallel_test.go` with C harnesses under `internal/native/testdata/parallel`. Further runtime hazards found during the audit will be claimed here before editing. No lowering or JavaScript files belong to this unit.

Audit additions to the runtime claims: `internal/native/library.go` (the actual runtime-library file), `internal/native/reuse.go` (one count-query hook), and `internal/native/runtime/string_share.c` (the same atomic count query). Shared counts cannot be read plainly by string sharing or generated reuse checks.

Additional runtime claims: `internal/native/runtime/region.c` initializes the slab/region marker used by sharing; `closure.c` will initialize result metadata once the ABI choice is agreed. Readonly Map iteration has an atomic iterator tally.

Final runtime ABI decision from @system_adamic_runtime: `adamic_parallel_map(items, work, bool references)` takes the result reference flag explicitly, just as `adamic_array_new` does. No closure result metadata is needed. Runtime implementation, C proofs and measured costs are recorded in [concurrency-runtime.md](concurrency-runtime.md); source lowering, the parallel oracle variant and the `.a` file benchmark are now integrated in [concurrency-integration.md](concurrency-integration.md).

Inline-count follow-up: `internal/native/release_test.go` updates the merged premature-free mutant to exercise both the new header dispatch and the last-reference decision in `heap.c`.
