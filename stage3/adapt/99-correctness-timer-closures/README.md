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

## Polling queue composition

Adamic and Compiler, task #0gj665k, Nov 14 00:59Z. The fixed-layout NodeArray
road in docs/rulings.md:76 does not admit PollingIntervalQueue.
internal/lower/node_array.go:nodeArrayLayoutOf keys on checker-resolved
declaration ancestry reaching the library Array/ReadonlyArray, a represented
numeric index element contract, and the actual extra properties. It does not
key on a coincidental type name. NodeArrayLayout then requires the fixed
pos/end/hasTrailingComma/transformFlags schema, primitive metadata and only
optional additional fields. PollingIntervalQueue lacks those required slots and
instead has three required polling fields, including pollScheduled: unknown.
nodeArrayCast also requires a direct array-literal operand and an equal mutable
element contract; the existing double cast is not that construction.

Main's full census retains the cast refusal at sys.ts:272:23 and five queue
value/return NotYets (three value sites and two function-return sites). Neither
the ruling nor the lowering implementation supplies a generic expando-array road.

Road (a) instead gives the queue a fixed object layout containing its files array.
files is (WatchedFileWithUnchangedPolls | undefined)[] because pollQueue writes
undefined and pollWatchedFileQueue already accepts that element contract.
pollScheduled keeps its adapted unknown type. Factory initialization is
{ files: [], pollingInterval, pollIndex: 0, pollScheduled: false }; the three
direct array uses select queue.files. Metadata, callback captures and the
existing helper's element guards stay as they were. A new reader requiring a
narrowing is a stop for Adamic to rule, not a reason to add a check.

This is JavaScript-changing, value-identical composition. The separate evidence
branch retains the complete use/reader inventory, the full census pair, the
236-event Node witness, the restoring-array mutant and the apply/types/lane
receipts.

## Changed-file queue element amendment

Adamic, task baobab, Nov 14, 2026. This amendment is type-only:
`changedFilesInLastPoll` is `(WatchedFileWithUnchangedPolls | undefined)[]`.
`pollQueue` writes undefined into its slots when unchanged files leave this
queue, and `pollWatchedFileQueue` already accepts that element contract. The
whole createDynamicPriorityPollingWatchFile after pin is amended here; its
upstream before pin and every other owner are unchanged.

The binding's uses are pollQueue, length, identity comparisons and push. No
reader needs a new narrowing or guard. The full single-file CommonJSBundle
sys.ts census removes only the invariant-mutable refusal at adapted 290:19,
with no new or changed finding. The restoring-declaration mutant brings that
refusal back. Stock TypeScript 6.0.3 emits byte-identical sys.js; adamic types
exits zero over all 79 compiler roots, and the whole-owner contract passes.

The complete census streams and delta, named restoring mutant, JavaScript
comparison, source hashes, apply and whole-lane receipts belong on
`typescript/changed-files-undefined-evidence`, under
`stage3/census/changed-files-undefined/`.

## Canonical Path result amendment

Adamic, #xn6fj0p, Nov 14, 2026. This amendment is type-only and follows
`typescript/scout-618` road (b): `toCanonicalFilePath` retains the string-returning
`createGetCanonicalFileName(useCaseSensitiveFileNames)` without a function cast.
The four Path-consuming calls mint their results with `as Path`: the directory
cache key, both recursive `invokeCallbacks` branches, and the child-watcher
cache lookup. The canonicalizer passed to `getRelativePathFromDirectory` and
the `includes` reader keep the string contract. No helper or normalization is
introduced. Only this owner's after text and SHA-256 change; the before pin
and the earlier timer and queue amendments retain their existing text.

The complete main/amendment single-file CommonJSBundle census delta, restoring
function-cast mutant, whole-owner contract, 79-root stock and Adamic checks,
per-file emitted JavaScript comparison, upstream `hereby dts` comparison, and
whole unfiltered lane receipts belong on `typescript/path-mint-618-evidence`,
under `stage3/census/path-mint-618/`.

## Retired poller control

The `poller` red control is retired: after `99-correctness-polling-no-enum-write`,
`pollingInterval` is an unreassigned parameter, the three queue bindings are
`const`, and no source, harness or test writes `PollingInterval` members.
The capture keeps tsc's argument evaluation order, but for this argument that
order is not observable while the queues are const and the enum is unwritten.
The label and `['/probe.ts']` queue-visit assertions remain behavior checks.
If a later change makes selection observable, a red control returns with it.
