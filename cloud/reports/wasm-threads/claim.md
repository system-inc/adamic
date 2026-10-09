# W7 claim

Branch: codex/wasm-threads, based on origin/wasm/integrate.

Inspect and, only where needed, add small ADAMIC_TARGET_WASI hooks in:
internal/native/runtime/parallel.c, adamic.c, adamic.h, count.c, count.h,
exceptions.c, heap.c, object.c, stack.c, string_index.c and weak.c.
Add the Threads section of docs/wasm.md and evidence under cloud/reports/wasm-threads/.

Concurrency integration candidate: origin/codex/concurrency at ced5530.
Async, moves and scaling are sibling extensions; none contains all others.
The concurrency integration report identifies the integrated pool/compiler oracle baseline.

Runtime hooks stay local and unpushed pending the user's runtime ownership clearance.
Only this claim, reports and the docs/wasm.md Threads section may be pushed now.
