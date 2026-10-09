# `nexus/correctness-require-child-process-error-listener`

| | |
|---|---|
| **Recommendation** | **Yes, but registered `off` until the sites are fixed.** 19 findings on ahra, 0 false. Each is one added listener |
| Findings | **ahra 19** (measured 2026-10-03) |
| Measured precision | 19 of 19 true: every reported child has no `'error'` listener anywhere in its file and is never handed away |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `spawn` and `fork` to `node:child_process`'s declarations; without a checker the rule registers nothing |

## What it checks

A ChildProcess returned by `node:child_process`'s `spawn` or `fork` that never gets an `'error'`
listener (`.on`, `.once`, `.addListener`, `.prependListener` or `.prependOnceListener` with `'error'`) and
never escapes: it is not returned, passed to a function (`events.once(child, 'exit')` counts, and it does
reject on `'error'`), stored in an object, array or property, exported, copied, or destructured. The
child is followed through chains of the emitter methods that return it (`.on('close', f).unref()`) and
into a local variable declared from or assigned the call, whose every reference in the file is read. The
doc comment carries the precise version.

Node emits a spawn failure (ENOENT, EACCES, EAGAIN, an aborted `signal`, a failed `kill`) as an `'error'`
event on the next tick, and an EventEmitter with no `'error'` listener throws it as "Unhandled 'error'
event", which kills the process that spawned the child.

## What it declines, and why

- **`exec` and `execFile` are not producers.** node's own `execFile`, which `exec` calls, runs
  `child.addListener('error', errorhandler)` on every child, with or without a callback (read from
  `lib/child_process.js` embedded in node 24.14.1). A missing binary there reaches the callback, or
  nowhere, and never throws. The task's precision line named `exec`; including it would make every
  `exec` without its own listener a false finding.
- **Placement in time is not judged.** A listener attached after an `await`, or on one branch only, is too
  late or partial, and still counts as handling the child. Any `'error'` registration anywhere in the file
  silences the variable. This can only miss a finding, never invent one.
- **Escapes are trusted.** A child returned or handed to a helper may get its listener there
  (`OllamaApi.ts:187` returns `{ process: proc }` to its caller). A receiver that does not attach one is a
  missed finding.
- **An event argument that is not a string literal type** (`string`, a spread) counts as `'error'`. A
  union of literals that excludes `'error'` does not.
- **`shell: true`.** A child spawned through a shell gets a missing *program* as exit code 127, not an
  `'error'`. It is still reported: the shell itself can fail to spawn (EAGAIN), `kill` can fail, and the
  child still has no listener. `FigmaMcpLauncher.ts:23` is the one such site.
- **A process-wide `uncaughtException` handler** is not a listener on the child and is not looked for.

## Where it came from

`modules/os/sensation/AhraOsMonitors.ts:140` in ahra: `prCheckStream` spawns `gh pr checks --watch` and
listens to stdout, stderr and `'exit'`, never `'error'`. A missing `gh` or an EAGAIN kills the whole
sensation daemon. Found by the cross-language pass of the new-rules sweep (`#tevhg3f`, after clippy's
`zombie_processes`), built as `#8gcxzsw`.

## Reconciling with the research count

Research counted **9** ("no listener and no escape"). cohere counts **19**. All 9 are in the 19 (two have
drifted lines: `PresenceApi.ts:222` is now `:226`, `PhiAnalyticsApi.ts:281` is now `:285`). The other 10,
none false:

- **+9, chained and unbound.** `NodeChildProcess.spawn('open', [path], { detached: true, stdio: 'ignore'
  }).unref()` at `TasksAttachmentsCommandLineInterface.ts:255`, `FinanceDocumentsCommandLineInterface.ts:276`,
  `ThingsCommandLineInterface.ts:421` and `MidjourneyApi.ts:543`, `:606`, `:634`, `:672`, `:736`, `:954`. The
  child is never stored, so nothing can ever attach a listener; the research probe ("discarded: 0")
  looked only at bound children.
- **+1, `fork`.** `FigmaMcpLauncher.ts:11` forks the socket server and only `.unref()`s and `.kill()`s it.
  The research probe counted `spawn`.

The research's one escape is `OllamaApi.ts:187`, silent here too.

## The findings, read one by one

All 19 read in context. Each child has no `'error'` listener and no escape in its file.

| Shape | Sites | Reading |
|---|---|---|
| Long-lived daemon child | `AhraOsMonitors.ts:140` | **Bug**: the sensation daemon dies with the child |
| Child awaited through a promise that resolves on `'close'`/`'exit'` only | `AhraOsBootCommandLineInterface.ts:218`, `OllamaCommandLineInterface.ts:19`, `:116`, `PhiAnalyticsApi.ts:285`, `PresenceApi.ts:226` | **Bug**: a failed spawn throws instead of settling the promise |
| CLI child that only listens for `'close'` | `PlanetScaleCommandLineInterface.ts:56`, `FigmaMcpLauncher.ts:23` | **Bug**: crashes with an unhandled event instead of a message |
| Detached child, unref'd | `OllamaCommandLineInterface.ts:65`, `FigmaMcpLauncher.ts:11`, the nine `open` sites above | **Bug**, low blast radius: a missing `open` cannot happen on macOS, but EAGAIN can, and it kills the command after its work is done |

**Recall spot check**: ahra has 40 `spawn`/`fork` calls from `node:child_process`. The 21 not reported
were each read: 19 attach `'error'` (`.on('error', ...)`, `.once('error', ...)`, or `on('error', reject)`),
`Run.ts:128` hands the child to `events.once(child, 'exit')`, which rejects on `'error'`, and
`OllamaApi.ts:187` returns it to its caller.

## Verification

- Fixtures both ways from the real sites: `AhraOsMonitors.prCheckStream` before (one finding on the
  spawn) and after (`child.on('error', ...)`, silent); `PresenceApi.getRecentConversationContext` through
  its `getBuiltinModule` cast, before and after. Shapes from `MidjourneyApi.ts:543` (chained `.unref()`),
  `AhraOsBootCommandLineInterface.ts:218` (promise on `'exit'` only), `:62` (listened and returned,
  silent) and `FigmaMcpLauncher.ts:11` (`fork`). 16 firing fixtures in all, 22 silent, plus the no-checker
  decline.
- Mutation check, each mutant a copy through `go test -overlay -count=1`, each killed: any listener
  counted as `'error'` (the monitor and presence fixtures and five more fire nothing); `exec` and
  `execFile` added as producers (the exec/execFile silent fixture fires); references matched by name
  instead of symbol (the same-name variable fixture and five more go silent); escapes read as neither
  (handed to a function, stored, a method handed on, now fire).
- `gofmt -l` and `go vet` clean through the overlay. The whole suite through the overlay is green apart
  from the two expected under an overlay: `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` (this rule
  is not in the live config yet) and `TestRulesFlagListsEveryRegisteredRule` (its subprocess builds without
  the overlay). A first-run timeout in `command/cohere` (`TestOneFixFormatRunSettlesAPrintedArrow`) passed
  on the rerun.
