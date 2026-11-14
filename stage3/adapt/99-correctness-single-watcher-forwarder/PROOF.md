# Proof and verification contract

The sole generic implementation cannot manufacture arbitrary T from a cast.
Instead it accepts a producer returning exactly T; each of the only two caller
instantiations constructs its own concrete function. A FileWatcherCallback
accepts (string, FileWatcherEventKind, Date | undefined); an FsWatchCallback
accepts ("rename" | "change", string | null | undefined, Date | undefined).
Each emitted arrow has three formal parameters and calls each subscriber with
those exact three values in order. No assertion or any annotation bridges the
callback types.

The getter runs inside each dispatch and obtains the current cache entry's
callback snapshot. Registration after the initial event is visible on the next
event. Removal during dispatch does not alter the existing snapshot. The helper
still owns cache creation, canonicalization, reuse, orderedRemoveItem by callback
identity and closing the last underlying watcher. The getter remains inside the
createWatcher argument, so synchronous delivery before cache insertion still
finds no callbacks. A missing cache entry does nothing. Both original and adapted
callbacks return undefined. Their formal length is three.

The witness uses the actual stock-emitted helper and entire enclosing factory,
exposing its two nested caller functions through the return object for testing.
It injects workers for OS delivery and exercises polling plus the real
fsWatchHandlingExistenceOnHost present-entry path. Both traces must be identical;
subscriber addition/removal, canonical cache reuse, stale close handles, close by
identity, timestamp identity, explicit undefined and function length are asserted.
The source/JavaScript scope check strips only the helper and the two nested
caller functions and compares everything else byte for byte.

contract.cjs verifies exact owner drift, moved/nested owners, duplicate owners,
half-adapted owners, a partial helper signature, invalid parsing and no-write
refusals. Both LF and CRLF trees accept unrelated additions. A second application
performs zero writes and preserves bytes and nanosecond mtimes.

All measured output belongs to the separate evidence branch
`typescript/single-watcher-forwarder-evidence`, directory
`stage3/census/single-watcher-forwarder`. Its commands.jsonl and setup command
record provide commands, environments and exits. The acceptance checks are:

- Full `stage3/apply.sh`, including all pre-existing adapters and this adapter.
- All-src textual and AST caller inventories: exactly two calls, zero others.
- The two identical Node witness traces and owner contract.
- `adamic types` over all 79 sorted compiler roots: zero diagnostics.
- Stock TypeScript 6.0.3 over those roots with the scout's strict options: zero
  new diagnostics; stock sys.js emission supplies the runtime witness.
- Full single-file sys.ts CommonJSBundle censuses before, adapted and restoring
  mutant with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE set.
  The original helper cast refusal must disappear and return in the mutant.
  Exact and unchanged-line-aligned ledgers list every new/gone/changed finding.
  The getter's optional slice call is #036ve6w's expected remaining stop.
- Whole `stage3/lane/run.sh`, unchanged runners, eight workers and no filter;
  its sanctioned declaration failure is the existing expected outcome.

The restoring artifact is named sys-mutant.ts. The census tree places those
bytes at src/compiler/sys.ts so the original imports and 79-root population
remain intact. Only the helper's createWatcher argument is restored to the
original unchecked `as T` expression; the typed factories remain in its callers.
The unused createForwarder parameter deliberately remains in that mutant.

The full census is measurement-only, returns no usable IR and is not a claim
that all of sys.ts lowers or that the native compiler runs TypeScript. OS error,
missing-entry, timestamp wrapping and polling fallback paths are outside this
bounded Node witness. No Loom submission is part of this work.
