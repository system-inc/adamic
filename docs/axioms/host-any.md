# The host's `any`, audited

Node's declarations are axioms too. Adamic bundles @types/node 25.3.3 and undici-types 7.18.2 (internal/load/node_types), and an `any` they hand a program is a value no one proved. This file gives each of those positions one disposition, the same three as the library's audit beside it (library-any.md). It is the only place the relation walk's trust in a host `any` may come from (ruled October 31, 2026 by @system_adamic on #p76cwjw; the audit is #c431sb6, held by @system_adamic_platforms_cloudflare_workers, and Language reads every modelled entry before it lands).

## Scope and method

The files are every `.d.ts` under internal/load/node_types/node_modules at main `dc8a6497`. @types/node's `ts5.6/` and `ts5.7/` folders are typesVersions redirects that TypeScript 5.8 and newer never load, so their 22 positions (all In) are left out of the counts below. Each `any` keyword is found by the TypeScript parser (6.0.3, `createSourceFile`), never by text search, and given a polarity by library-any.md's rule: a parameter flips it, and a parameter of a callback flips it back. A writable property that hands a value out is Both.

| Polarity | Positions |
| --- | --- |
| Out: a return type, a property, a type argument in one, or a parameter of a callback Node calls | 303 |
| Both: a property the program can also write | 42 |
| In: a parameter of a Node function | 284 |
| Defaults (74) and constraints (104) of type parameters | 178 |
| Patterns in a conditional type's `extends` clause | 140 |
| **Total** | **947** |

Value-carrying means Out or Both: 345 positions. 210 of them are the `...args: any[]` of an event listener (`on`, `once`, `addListener` and the rest, across Node's event emitters).

## Why most of the host is counted already

Lowering admits a Node member only when it's registered with a real lowering and runtime (`implementedNodeMembers`, internal/lower/library_node.go:13, filled by `RegisterNodeLibraryMembers`). Any expression whose symbol is declared in the bundled Node library and isn't registered is NotYet, named, in the up-front refusal pass (`nodeLibraryRefusal`, library_node.go:85-87, refusal rule 800). That pass runs before lowering, so before any relation. Every `any` in an unregistered member is therefore counted: the program never receives its value. This holds for the listener callbacks too, since the methods that would call them (`process.on` and the rest) aren't registered.

The registered set at `dc8a6497` is 99 names (the console registry's three plus `RegisterNodeLibraryMembers` in library_node.go, library_node_fs_buffer.go, library_node_fs_directory.go, library_node_fs_file.go, library_node_host.go and library_node_process.go). Two more are bound by name without registration, both read only on `process.stdout`: `node:tty.WriteStream.isTTY` and `node:tty.WriteStream.columns` (library_node_host.go). Each `any` was joined to its member by the name `nodeLibraryMember` builds for its declaration.

**Value-carrying `any`s inside a registered or bound member: none.** So this audit has no Modelled entries yet, and the relation walk's host list is empty.

## Corrected

None for `any`. Host corrections that only add truth live in internal/load/host_axioms.json (Lower's loader table, landed 851bb41a), served in place of the published line. Its rows today correct `boolean` and `number` members that are absent off a terminal (tty's isTTY, isRaw, columns, rows and four methods, #jk7tph7). A future `any` correction goes there, counted on tsc first, with a row naming its ruling and Node's measured value.

## Modelled

None. A registered member whose declared type carries a value-carrying `any` gets an entry here before it's registered: the call, the lowering rule that supplies the real type, its adapter's argument and result checks, and a pin that goes red with the rule removed (library-any.md's standing rule, from M3). Language attacks it before it lands.

## Registered members whose `any` carries no value

| Declaration | Where | Polarity | Why no value reaches the program |
| --- | --- | --- | --- |
| `Console.log(...data: any[])`, `Console.error(...data: any[])` | console.d.ts:107, :102 | In | The program hands values to the console. Nothing comes back |
| `Writable.write(chunk: any, ...)`, both overloads | stream.d.ts:900, :901 | In | Admitted only as `process.stdout.write(s)` with one string argument, as a discarded statement (#grqdgpr). A non-string chunk is refused by name |
| `Process.nextTick(callback: Function, ...args: any[])` | process.d.ts:1827 | In | Node would pass `args` to the callback, which would carry them back in, but a call isn't admitted. Only the feature test `!!process.nextTick` and `typeof` lower (library_node_process.go:341, "only the Node-like feature observation, not callback scheduling") |
| `performance`'s `onmessage: any` in a conditional's `extends` clause | web-globals/performance.d.ts:44 | Pattern | A shape the conditional matches, never a value |

## Counted

Every value-carrying `any` in an unregistered member, by declaring module. Each is NotYet at the member before any relation (library_node.go:85-87).

| Module | Value-carrying |
| --- | --- |
| node:http2 | 56 |
| node:stream | 49 |
| node:http | 32 |
| node:fs | 21 |
| node:events | 20 |
| node:test, node:tls | 17 each |
| node:worker_threads | 15 |
| node:repl | 11 |
| node:process | 10 |
| node:https, node:module, node:net, node:perf_hooks, node:tty, node:vm | 8 each |
| node:inspector | 7 |
| node:inspector/promises, node:stream/web, undici-types/websocket | 5 each |
| node:cluster, node:util/types, node:v8, undici-types/eventsource | 3 each |
| node:async_hooks, node:iterators, node:util | 2 each |
| node:abortcontroller, node:domain, node:globals, node:storage, node:wasi, undici-types/diagnostics-channel, undici-types/dispatcher, undici-types/errors, undici-types/mock-interceptor | 1 each |
| **Total** | **345** |

Twelve of them are reachable as globals when a program uses Node's types without an import, and they are counted the same way, since their symbols are declared in the bundle and none is registered:

| Global | Where | Polarity |
| --- | --- | --- |
| `module.exports` | module.d.ts:652 | Both |
| `module.require(id): any`, `require(id): any` | module.d.ts:700, :707 | Out |
| `performance.toJSON(): any` | perf_hooks.d.ts:116 | Out |
| `process.on`, `once`, `off`, `addListener`, `prependListener`, `prependOnceListener`, `removeListener` and `listenerCount`: the listener's `...args` | process.d.ts:2102 to :2146 | Out |

## The list the walk reads

The relation walk's host list is the Modelled section's ids, and today it's empty. The walk carries no host exemption of its own, so every host `any` it meets is counted. Registering a Node member is the point where this file has to change. Before a registration whose declared type carries a value-carrying `any` lands, the `any` gets a Modelled entry with its pin, or a correction in host_axioms.json. Otherwise it stays counted, and the program fails closed at the relation.
