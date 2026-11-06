# The host bindings, audited

A `.ts` file may declare a value the host supplies, as tsc does: `declare let onProfilerEvent` in performance.ts, `declare const process` in core.ts, `declare const performance` in performanceCore.ts. docs/rulings.md refuses a program's `declare` of a value and trusts only Adamic's own declarations, each an axiom under Language's audit. Such a declaration is admitted only where the host-binding table names it, and the table is one of those axioms (ruled by @system_adamic on #00qn5az, Nov 5 and 6, 2026; the ledger's entry "A `.ts` file's ambient value declaration is a host binding"). This file is its audit, beside host-any.md, and Language reads every row before it lands.

The table is internal/load/host_bindings.json. TestHostBindingsAudit (internal/load/host_bindings_test.go) holds the rows below to it both ways, cell for cell, and fails on a row with no ruling.

## What a row says

- **Declaration**: the file's basename, the keyword and the declared type's exact text. Only a top-level, unexported `declare` statement in a `.ts` file matches, every declarator of it a row's. Anything else stays refused as an ambient value declaration, and so does every `.a` file's, which is otherwise unchanged. A program's `declare const process: number` is never the row's.
- **Node**: what Node prints for `typeof` of the name, on the version named, running the declaring file as an ECMAScript module, as oracle/node.mjs runs every program. Under that module mode `typeof require` is "undefined", which is why tsc's `require` leaves the program by adaptation rather than by a row.
- **native, javascript, wasm32**: Provided (the host defines the name, and reads go through its host model, never trusted), Absent (nothing defines it, and `typeof` answers "undefined"), or Pending with the owner who clears the column. Runtime clears the native column, and Runtime and Platforms clear the wasm32 one with it.
- **Model**: the host path a Provided row's reads route to, or the task that builds it while none is built.
- **Divergence**: a ruled backend divergence under which a column may differ from Node's answer, or none.
- **Ruling**: the ruling or task that made the row.

## What a row admits

Lowering makes one IR for every target, so a row lowers only when its three columns are equal and none is pending. Any other row is not yet, named, at its declaration and at every use (internal/lower/host_bindings.go, `hostBindingStop`).

An Absent row's `typeof` is the constant "undefined". The name itself is never read. A read is admitted only in the then-branch of `if`, the right of `&&` or the true arm of `?:`, under a guard `typeof X === "t"` with t not "undefined", or `typeof X !== "undefined"`, matched to the row's declaration by symbol. A shadowing local is judged as the local. Every other shape is not yet: `||`, an early return, a negated guard, a guard bound to a local, a guard in a called function. The guarded code is never lowered. It becomes a compiler-bug panic, reached only if the guard's test was wrong. A write to the name, and a `globalThis.<name>` or `globalThis[name]` write naming any row, are refused.

Every column is held to Node's real `typeof` by internal/oracle/host_bindings_test.go, whether or not the row lowers: Absent where Node prints anything but "undefined" fails, and so does Provided where Node prints "undefined", unless the row names a ruled divergence.

## Rows

| Name | Declaration | Node | native | javascript | wasm32 | Model | Divergence | Ruling |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| onProfilerEvent | performance.ts, `declare let onProfilerEvent: ((eventName: string) => void) \| undefined` | v24.21.0, ES module, `undefined` | Absent | Absent | Absent | none | none | @system_adamic on #00qn5az, Nov 6 2026 00:27Z: Absent on every target, Node's own answer under the oracle's module mode |
| process | core.ts, `declare const process: { nextTick?: unknown; browser?: unknown; } \| undefined` | v24.21.0, ES module, `object` | Pending (Runtime) | Provided | Pending (Runtime and Platforms) | `node:process` | none | @system_adamic on #00qn5az, Nov 6 2026 00:05Z: Provided through Node's process model, the native and wasm32 columns pending Runtime's and Platforms' clearance of process.browser |
| performance | performanceCore.ts, `declare const performance: Performance \| undefined` | v24.21.0, ES module, `object` | Pending (Runtime) | Provided | Pending (Runtime and Platforms) | #3sh9398, not built | none | @system_adamic on #00qn5az, Nov 6 2026 00:05Z: Provided on Node, not yet until #3sh9398's host model exists |

Only onProfilerEvent lowers today. process waits on Runtime and Platforms to clear `process.browser`, which Node answers undefined, before its reads route through the Node process lowering (internal/lower/library_node_process.go). performance waits on those columns and on #3sh9398's host model.

## Changing a row

A row changes here and in the JSON together, with the ruling that changed it. A column moves off Pending only when its owner clears it on #00qn5az or a task that names this file. A Provided column's model has to exist before its row lowers.
