# Timer closures: baseline metadata blocker

Task 995spev, Adamic's Nov 13 05:07Z and 05:31Z rulings. The user supplied
those comments verbatim and they were read before this probe.

Base main: e3f59bf5509b1feef85b405807a732e11dc383b6.
Count: a3a5bb705467dd4473de73a0e1a41ac7f0d45858.
Model: 66e4051f19854166c95df6be99a4ae0e2eeaadd8.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8, version 6.0.3.

## Observed blocker

The callback bodies can preserve their behavior, but the timer tails have a
second observer. TestServerHost.setTimeout forwards the actual tail array to
Callbacks.register. Callbacks.log reads args[0], and switchToBaseliningInvoke
also reads args[0]. These values appear in existing non-API baselines.

A closure supplied with only callback and ms gives TestServerHost an empty
tail array. The unchanged register-and-invoke path then logs undefined instead
of the original label. Making the callback signature concrete does not restore
that information. Leaving the harness method variadic also does not restore it.

The executable counterexample extracts the complete actual dynamic poller,
complete actual Callbacks class and actual TestServerHost.setTimeout method.
Only the minimal host clock and initial file stat are supplied. The actual
poller schedules a file, and the actual callback registry logs its pending
state. No timer has to fire to observe the difference. The replacement captures
the assignment receiver, timer callee, callback, delay and both tails in their
original order, preserving the timer receiver with call. This isolates the
baseline metadata problem from late evaluation or dynamic-this problems.

Case timer-label-baseline:

```
before: Timeout callback:: count: 1\r\n1: pollLowPollingIntervalQueue
after:  Timeout callback:: count: 1\r\n1: undefined
```

The byte-identity assertion exits 1, by design. This is a focused counterexample,
not a complete lane run or a completed adaptation. result.json was collected
while main's apply was still running; upstream-result.json independently
reproduces the same mismatch against the immutable pinned upstream checkout.

For a real checked-in baseline example, see TypeScript's
`tests/baselines/reference/tsserver/watchEnvironment/uses-dynamic-polling-when-file-is-added-to-subfolder.js`,
lines 263 and 264:

```
7: pollLowPollingIntervalQueue *new*
8: pollPollingIntervalQueue *new*
```

The probe requires a decision about an explicit harness metadata path if those
bytes must remain identical. Such a path has not been implemented or presumed
sanctioned here. No API sanctions, baseline expectations, ambient declarations,
or adaptation owner pins have been changed to conceal the mismatch.

## Reproduce

From Adamic main with the locked stage3/api dependencies installed:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node \
  stage3/census/timer-closures-blocker/probe.cjs \
  /workspace/scratch/typescript-source
```

Expected child exit: 1. Raw JSON: upstream-result.json. The probe directly sets
its own failure exit code, without a wrapper that replaces the child's result.
The script asserts both exact observed strings before reporting the byte mismatch.

## Owner-pin findings before edits

- getNodeSystem is pinned whole by 93-builtin-imports and
  99-wasm-crypto-import. Any changes inside it must amend its existing adapters.
- 92-tty-host-results pins its two terminal methods inside getNodeSystem.
- createWatchProgram is pinned whole by 99-correctness-axolotl-x. Its two timer
  wraps would be amendments there. 97-cli-axolotl-c also edits its close path.
- 41-explicit-any-remaining owns the ambient timer pair. Neither declaration
  was deleted or edited.
- The initial hits in 99-correctness-axolotl-q and -r were references to System
  inside other owners, not ownership of the timer functions.
- No amendments were made. The remaining owner map was not certified complete.

## Work deliberately not represented as proved

No adaptation commit, slot declaration changes, ten-call production rewrite,
API sanction regeneration, adapted replay, Adamic type proof, 79-root diagnostic
comparison, complete lane, three required temporal witnesses, three required
mutants, ledger change, or final sys.ts latent census exists at this stop.
Only the focused baseline counterexample is claimed. It is not one of the
three requested production mutants.
