# Correctly typed single-watcher forwarders

Road C from Adamic's ruling on #1vkqj9w (Nov 14, 09:09Z). The shared
createSingleWatcherPerName implementation accepts
`createForwarder: (getCallbacks: () => T[] | undefined) => T`. Its polling and
filesystem callers each construct their concrete callback, with three formal
parameters, forwarding all three values in order. The getter reads and snapshots
the current callback list at dispatch, retaining registration, removal, cache
reuse and close-by-subscriber-identity behavior.

This uses scout C, including its local `callbacks` guard, rather than C-minimal.
Their callback values, snapshot behavior and undefined return are identical;
the scout measured identical census findings. C is the requested reviewed edit.
The emitted JavaScript changes because code moves from the generic helper to
the two callers. Value identity covers the tested event and cache contract;
function source reflection and allocation timing are outside that claim.

The initial owner search covers createSingleWatcherPerName, pollingWatchFile,
fsWatch and enclosing createSystemWatchFunctions throughout stage3/adapt.
99-correctness-timer-closures owns different compiler functions and harness
classes mentioning createSystemWatchFunctions; it does not pin either edited
compiler owner. Therefore this is a new adaptation, not an amendment.

owners.json pins the entire helper and the entire enclosing
createSystemWatchFunctions, including both call-site functions, as exact text
and SHA-256. The adapter parses before planning and after replacement, rejects
drift, duplicates and mixed states before writing, rereads before writing,
preserves CRLF and makes no writes on replay. contract.cjs exercises those gates.

resolved-sites.json resolves adaptation 40's two declined rows, identified by
file, line, column and source: src/compiler/sys.ts:516:30 and :516:58. Both any
annotations disappear; there is no replacement any or cast. The repository
search found no executable reader of 40's historical evidence/sites.json
dispositions, so that historical inventory is unchanged.

Evidence is on branch typescript/single-watcher-forwarder-evidence, under
stage3/census/single-watcher-forwarder, separate from this directory. It includes
the exact owner and caller searches, commands/exits, emitted source comparison,
both Node traces, all full census streams and every exact/source-aligned finding
delta. The optional call in the snapshot getter remains #036ve6w's concern;
it is recorded rather than changed.

Run with NODE_PATH pointing to stage3/api/node_modules:

```
node stage3/adapt/99-correctness-single-watcher-forwarder/adapt.cjs TREE
node stage3/adapt/99-correctness-single-watcher-forwarder/contract.cjs BEFORE_TREE
```

The runtime witness evaluates stock-emitted helper and complete watch factory,
with only its return object instrumented to expose the two actual nested caller
functions. Injected polling and filesystem workers replace OS event delivery;
fsWatchHandlingExistenceOnHost still executes its actual present-entry path.
No claim about all OS event/error/fallback paths follows from this witness.
