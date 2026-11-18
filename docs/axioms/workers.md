# The Workers bindings, audited

A Worker's environment can hold a Workers binding, and the prelude's `adamic/workers` module (internal/load/prelude.d.ts) declares what a program may call on one. Those declarations are axioms: Workers' own runtime answers the call, and nothing Adamic compiles proves the declared type. This file gives each admitted call one row, its premise, and the guard that holds the value the binding hands back (#bxaz0p1, head 3b-1; Language audited the get row before it landed). It sits beside host-any.md and host-bindings.md, and Language reads every row before it lands.

## What a row admits

Lowering admits a Workers member only on the road its row names, and every other use of a binding stays rule 149's NotYet (internal/lower/library_workers.go). The bridge checks each KV binding before the handler runs (environmentKVValidation, internal/worker/environment.go), but that check proves only that the binding's methods are functions, never what they return. The guard in a row's last column is what holds the value.

## Rows

| Declaration | Road | Premise | Guard | Pins |
| --- | --- | --- | --- | --- |
| `workers:KvNamespace.get(key: string): Promise<string \| null>` | The call is the direct operand of `await`, parentheses allowed, on a binding read as an identifier or a property access (workersAwaitedGet). Lowered to ir.NodeHostCall, Module `workers:KvNamespace`, JavaScript only; native refuses it | A missing key resolves to null, and with no type argument the value is text. The real binding also takes `get(key, type)` and `get(key, { type, cacheTtl })`, which resolve to an object, an ArrayBuffer or a ReadableStream for `"json"`, `"arrayBuffer"` or `"stream"`. Only prelude.d.ts's one parameter keeps a program from passing them, so widening that signature reopens this row | adamicWorkersTextOrNull (internal/worker/runtime.mjs) on every admitted get: the JavaScript backend emits `.then(adamicWorkersTextOrNull)` on the call itself (internal/javascript/library_workers.go), so no path to the await can skip it. Text and null pass unchanged; anything else is a panic naming workers:KvNamespace.get and the value's typeof, never the value | TestWorkerKVGetAwaitedUnderNode and TestWorkerKVGetThroughHelperUnderNode (a get that resolves 42 is the bridge's 500); TestWorkerKVGetTypeArgumentFailsToLoad (get_json.a stops at tsc's TS2554); the kv-text-or-null mutant in internal/worker/kv-capability.mutants.md |

A row that admits a second road, or a second member (put, delete, a typed get), is a new row here first.
