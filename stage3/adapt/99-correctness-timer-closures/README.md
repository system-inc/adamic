# Concrete timer slots and captured tails

Adamic, task 995spev, Nov 13, 05:07Z, 05:31Z and the 05:48Z label amendment.
This is JavaScript-changing. It changes the earlier C-noinfer-b ruling explicitly:
there is no generic slot, no NoInfer, and no TypeScript 5.4 reader requirement.

```
setTimeout?(callback: (label?: unknown) => void, ms: number, label?: unknown): unknown;
clearTimeout?(timeoutId: unknown): void;
```

System and public WatchHost retain optional methods. ServerHost's own required
redeclarations remain required. The internal polling picks and
RecursiveDirectoryWatcherHost follow System through NonNullable indexed access.
TestServerHost and SessionServerHost declare the concrete callback and label;
TestServerHost still registers a callback plus an argument array and invokes it
through the unchanged Callbacks.register/invoke path. Its sole label is stored
as [label]. SessionServerHost still does not invoke callbacks and returns undefined.
The typed mockHost's no-op suppliers already accept the concrete contract.

Every original tail becomes a schedule-time const. The zero-parameter firing
closure calls the captured original callback with those values. The first tail
is also passed as the optional label: it retains the exact value and identity
Callbacks logs. callbacks may truthfully observe that optional label, although
all ten supplied closures ignore it. Node and TestServerHost both forward it.

The expressions are replaced by immediately invoked blocks to preserve their
position inside surrounding assignments and calls. Each block evaluates the
original receiver first, captures the callback, evaluates the delay, then
captures all tails left to right. The nine receiver-based blocks then read and
call timerHost.setTimeout. WatchHost and the solution builder keep their original
presence guards and assert timerHost.setTimeout! at the receiver call. There are
no optional calls. The recursive watcher already receives a destructured local
and still calls it directly.
The IIFE does not defer any schedule-time read until firing.

The two static callbacks use their explicit self parameters. The session and
typing installer callbacks are arrows with lexical this. The other six are local
function declarations and do not read dynamic this. Function declarations that
ignored the original label gain an unused optional string parameter so the
explicit closure call states the original argument truthfully.

## Owners and amendments

owners.json pins all thirteen new owners whole, before and after. Exact text,
SHA-256, unique top-level ownership, parse validity, and uniform before/after
states are checked before any write. A changed, missing, duplicated, moved or
partly adapted owner refuses without writes. Replay writes nothing and preserves
bytes and mtimes. LF and CRLF are retained. Unrelated source is accepted.

Earlier owners are amended in place:

- 41 now deletes both ambient declarations atomically under road 5.
- 93 owns the two concrete setTimeout/clearTimeout forwarding pairs inside
  getNodeSystem. Its after pin and the crypto adapter's before/after pins change.
- 99-wasm-crypto-import retains its crypto edit, with amended whole-owner pins
  including 93's forwarding closures.
- 99-correctness-axolotl-x performs the two WatchHost wraps and optional unused
  callback parameters inside its whole createWatchProgram owner.

api-sanctions.json declares six emitted public member differences: System,
WatchHost and ServerHost each redeclare the pair. An external caller may pass one
label, but further tail arguments stop type-checking. A variadic implementer still
fits. The lane manifest, runtime hash and sanction count move together.

calls.json gives every original call and its replacement, keyed by main's adapted
coordinates. PROOF.md describes the checks and separate evidence destination.

## Premises a reader can re-check

Node calls a timer callback with the Timeout object as `this`. The firing arrow
now calls `callback(t1, ...)` with `this` undefined. This is sound only because
none of the callbacks scheduled through these 10 calls reads dynamic `this`:
the six function declarations, the two static methods (which take self
explicitly), and the two arrows (whose `this` is lexical). The count at
`typescript/timer-slot-count-995spev` (`a3a5bb70`) shows it. If upstream changes
one of those callbacks, re-check this.

The unused optional parameter added to watchPublic's two callback declarations
(`_timeoutType?: string`) changes each function's `.length` from 0 to 1. Nothing
in tsc reads a callback's `length`. Re-check this if that changes.

## Road 5: global cancellation parameter

Adamic, task 995spev, Nov 13 07:51:37Z: 41 deletes the ambient timer pair;
93's getNodeSystem cancellation forwarders use
`clearTimeout: (timeoutId: Parameters<typeof clearTimeout>[0]): void => { clearTimeout(timeoutId); }`.
The cancellation call is a statement in a block body; it does not return the
global call's result. The scheduling forwarders still return the handle, and
System's `unknown` method slot stays unchanged.
Stock tsc's acceptance rests on method-parameter bivariance at the System slot;
Adamic's relation is exact because its existing global provider accepts `unknown`.
A property-style slot (`clearTimeout?: (timeoutId: unknown) => void`) would fail
TS2322 in the lane. Re-check this premise if the slot declaration changes.

Evidence and commands are on `typescript/timer-ambient-delete-evidence-3`, under
`stage3/census/timer-ambient-delete-3/`, including both 79-root checks, the
restored-ambient and property-slot mutants, the Node cancellation witness,
full apply and replay, 398 shards, and the full-mode census pair.

## Receiver-call amendment

Adamic and Compiler, task #0gj665k, Nov 14 00:59Z; implementation #5vc7nbf.
This is JavaScript-changing: the nine blocks move the timer member read after
the argument captures. No capture, callback, delay, tail or presence guard changes.
The unused destructured method locals are removed.

The reorder is admitted only after both counts are zero across source, harness
and tests and in the whole lane oracle: timer-host setTimeout/clearTimeout
accessors, and argument expressions able to write either receiver method.
The audit, coverage limitations and actual run results are retained with the
separate evidence.

Proof streams, every file's full census delta, the mutant red/green fixture,
Node witnesses and lane receipts belong on typescript/timer-call-queue-evidence,
under stage3/census/timer-call-queue/.
