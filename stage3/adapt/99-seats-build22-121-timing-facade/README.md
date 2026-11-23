# 99-seats-build22-121-timing-facade

Family 121 on #bbssdcy follows Adamic's Nov 22 ruling at about 22:58Z and her Nov 23 answer at 07:21Z. Adamic accepts after review and census. The host-binding and runtime-shape rules are in `docs/rulings.md:230` and `:407`. The refusal is at `src/compiler/performanceCore.ts:74:33` in the before snapshot, inside `tryGetPerformanceHooks`.

The adapter reads `performance.timeOrigin` once into the local `timeOrigin` immediately before the guard. The guard checks the local's type and the timing slot copies its value. The slot's `now` closure calls `performance.now()` with the host as its receiver. The separate event slot still receives the host object. The adapter pins the local, guard and assignment by exact syntax-tree text.

Adamic’s ruling on #bbssdcy, Nov 22 at about 22:58Z, from `tail-proposals-2230.md`:

> "Compiler accepts" below means under the split; "mine" means I read the head.
>
> **Build:**
> - **121** (1): yes, the facade, mine. Proof owed first: HR-Time fixes `timeOrigin` for a realm's life, so the facade copies it once; show our host's `performance.timeOrigin` reads one value across a wait, natively and on Node, and that nothing in tsc writes or rereads it expecting change.

Adamic's Nov 23 answer:

> **9. Read once,** into a local, keeping the receiver-bound closure. performanceCore.ts:35 shows tsc means to accommodate polyfills.

The polyfill witness supplies a getter that yields 2718 and throws on a second read. The before module reads it once in the guard and retains the host in the timing slot. The adapted module initializes after one read and stores 2718 in the timing facade. This preserves the accommodation for polyfills at `performanceCore.ts:35`.

The governing rows are in the repository at the lines cited above.

## One clock origin across a wait

The HR-Time [timeOrigin getter](https://www.w3.org/TR/hr-time-3/#dom-performance-timeorigin) returns the time origin timestamp for its relevant global object. The origin is defined for that global, independently of the progression measured by `now()`.

At the pin, `internal/native/library_node_process.go:13` maps the member to `adamic_node_time_origin`. `internal/native_runtime/c/node_process.c:65` returns `adamic_host_time_origin()`. The host object's materialized slot uses that same value at `node_process.c:450`. `internal/native_runtime/c/host_runtime.c:67` declares the startup constructor `capture_clock_origin`. It pairs a realtime clock sample with a monotonic origin and assigns `epoch_origin` at line 78. This is the sole assignment. The getter at line 120 returns that stored number, while lines 114 to 118 compute elapsed time from a fresh monotonic reading.

The round 1 lane audit's `tmp-audit/time-origin.ts` imports Adamic's modeled Node host export and waits for 250 ms of measured monotonic time before the second read. Node v24.19.0, Adamic JavaScript on that Node, and unsanitized Adamic native (Clang 20.1.8) each exited 0 and printed exactly `true\ntrue\n`. The first boolean checks that the two origins are equal. The second checks that `performance.now()` advanced. The runner exited 0. See round 1's probe table and complete outputs, and review-j's rerun with the round 2 local, guard and shorthand store.

## Checker references before this adaptation

Round 1's lane audit includes `tmp-audit/reader-trace.cjs` and its trace for the tree immediately before family 121. The script builds a program rooted at all TypeScript sources of that input. It collects each spelling of `timeOrigin`, resolves symbols through the checker, and asks the language service for references to those symbols and the slot and its aliases. The full stage3 apply succeeded. The checker program covers 713 source files, including 38 in `src/harness`. The lane audit contains the 77 reference and spelling records, with the exact source line and the checker or syntax classification. The table uses before-snapshot coordinates. The facade moves the guard's host read into the local initializer and keeps the separate event read.

| Site | Source | Use |
| --- | --- | --- |
| `src/compiler/performanceCore.ts:17` | `readonly timeOrigin: number;` | The internal `PerformanceTime` member, inherited by `Performance` at line 21. |
| `src/compiler/performanceCore.ts:73` | `if (typeof performance.timeOrigin === "number" && typeof performance.now === "function") {` | Checks availability and type once during initialization. |
| `src/compiler/performance.ts:184` | `timeorigin = perfHooks.performance.timeOrigin;` | Copies the host origin into the compiler’s own lowercase baseline on enable. It does not compare origins between enables. |
| `src/lib/dom.generated.d.ts:27450`, `src/lib/webworker.generated.d.ts:8152` | `readonly timeOrigin: DOMHighResTimeStamp;` | Ambient library declarations. |

The original object comes from `const { performance } = perfHooks;` (core line 36) or the browser binding returned as `performance,` (line 55). `const p = tryGetPerformance();` (line 63) and `const { shouldWriteNativeEvents, performance } = p;` (line 65) carry the object to the store. At lines 74 and 84 the stores are `hooks.performanceTime = performance;` and `hooks.performance = performance;`. The former is checked for presence at line 78. The returned container is cached by `const nativePerformanceHooks = tryGetPerformanceHooks();` (line 90), and its timing field is read into `const nativePerformanceTime = nativePerformanceHooks?.performanceTime;` (line 91). Line 103 uses this alias for presence and the receiver-held `nativePerformanceTime.now()`. No caller compares the timing slot's identity with the event slot.

`return nativePerformanceHooks;` (line 95) exposes the containing object to its sole caller in `performance.ts:182`, `perfHooks ||= tryGetNativePerformanceHooks();`. That caller reads the event slot at lines 183, 184 and 190. The alias `performanceImpl = perfHooks.performance;` at line 190 is used for `mark`, `measure`, `clearMeasures` and `clearMarks` at lines 79, 103, 149 and 162, and reset to `undefined` at line 207. The write at line 184 targets the lowercase number, which is initialized at line 62 and used as a measurement fallback at line 100.

There are zero writes to a `timeOrigin` property and zero reads that expect a changed value. No runtime reads or writes were found in `src/harness`. Both proof requirements held before the adapter edit.

## Production adapter witness

The witness tests the original and adapted modules in fresh CommonJS contexts with the stock TypeScript 6.0.3 transpiler and fake Node and browser hosts. It checks the clock at 10 and 20, the host receiver, event identity and flags, the timing slot, absent capabilities and the effectful polyfill. It invokes the adapter beside the witness on a temporary copy of `performanceCore.ts`. A closure that returns the origin fails the clock row. A second origin read fails the polyfill row. The fixtures use origins 1000 and 2718, and a store that writes the literal 1000 fails the browser and polyfill rows.

Run from the Adamic checkout with a cold applied tree that holds the adaptation refs. Set `adapted` to that tree's path. Extract the adapter's before snapshot, which includes the earlier `node:perf_hooks` import. Give the witness a new output file outside the adapter directory.

```sh
adapted=/path/to/applied/tree
snapshot=$(mktemp -d)
git -C "$adapted" archive refs/stage3/99-seats-build22-121-timing-facade-before src/compiler/performanceCore.ts | tar -x -C "$snapshot"
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-121-timing-facade/witnesses.cjs "$snapshot" "$snapshot/verdicts.json"
```

## Adapted-tree byte guard

The adapted-tree guard compares lines 62 to 89 of the applied `tryGetPerformanceHooks` with `internal/lower/testdata/host_binding/performance_core_timing_facade.txt`, including the CRLF line endings. Lower's host-view tests continue to use the pre-facade body in `performance_core_hooks.txt`.
