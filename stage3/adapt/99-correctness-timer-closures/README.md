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
original receiver and timer member first, captures the callback, evaluates the
delay, then captures all tails left to right. A captured method is invoked with
Function.call and its original receiver. WatchHost and the solution builder use
a non-null assertion on that captured method under the original presence guards.
There are no optional calls. Nine calls use a destructured local's call method;
the recursive watcher already receives a destructured local and calls it directly.
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

- 41 retains both ambient declarations and gives them the concrete optional-label
  contract. Their deletion remains a separate stacked change.
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
