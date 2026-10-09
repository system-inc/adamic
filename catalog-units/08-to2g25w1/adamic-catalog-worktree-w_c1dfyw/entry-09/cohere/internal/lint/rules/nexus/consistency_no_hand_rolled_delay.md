# `nexus/consistency-no-hand-rolled-delay`

| | |
|---|---|
| **Recommendation** | **Yes** (option C of the `no-promise-executor-return` ruling, 2026-10-01: every sleep is `delay`) |
| Findings in ahra | **13** at HEAD on 2026-10-01; **2** in the working tree |
| Measured precision | 13 of 13 at HEAD, read by hand |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | no |

## Name

`consistency-` because the judgment is one spelling for one act, and `no-hand-rolled-delay` because
what is reported is a reimplementation of a named primitive. The review's working name was
`prefer-nexus-delay`; the nexus rules here are named for what they forbid.

## What it checks

`new Promise(executor)` whose executor does nothing but resolve after a timer, in favour of
`await delay(milliseconds)` from `@nexus/source/coordination/Delay`. Every condition must hold, and
each one is a way the promise could be doing something `delay` does not:

1. Callee is the identifier `Promise`, with exactly one argument (type arguments allowed).
2. The executor is an arrow or function expression with one or two plain identifier parameters; a
   second (`reject`) only when the body never mentions it. A rest parameter does not count.
3. The body is only the timer call: an expression body, or a block of one expression statement or
   one `return`.
4. The timer is `setTimeout`, `globalThis.setTimeout` or `window.setTimeout` (no optional chain:
   `window?.setTimeout` never settles where `window` is undefined), with exactly two arguments.
5. The callback is `resolve` itself, or a parameterless function whose whole body is `resolve()`
   with no arguments.

## The definition of `delay` is not a use

`delay` is written in this shape, and it does not report, with no path or name in the rule. A
promise that is the whole body of a function (its single `return`, or an arrow's expression body)
and whose duration is a bare reference to one of that function's own parameters is the definition
of a delay primitive. A fixed duration, a computed duration, a conditional return or any other
statement in the function makes it a use and it reports.

What this leaves unreported is a second definition elsewhere (`function sleep(ms) { return new
Promise(...) }`), a duplicate helper whose callers already read as a named wait. Measured at HEAD on
2026-10-01: the only function of that shape in ahra is `delay` in
`nexus/source/coordination/Delay.ts`.

## Silent on purpose, checked against the tree

`AirPlayRtspConnection.ts:131` keeps the timer for `clearTimeout` and resolves early on a reply.
`FrameTvWebSocketConnection.ts:154` uses the timer as a fallback beside `close()`.
`PentairApi.ts:61`, `CodexAppServerClient.ts:445`, `FrameTvApi.ts:184` and
`PhiSocialGenerator.ts:573` are timeouts that reject, not waits. `BackoffTask.ts:110` is a cancellable
wait that stores its timer and `resolve` so `cancel()` can cut it short. The `PromiseBarrier`,
`PromiseGroup` and `TrackedPromise` tests resolve with a value after a timer. `Delay.ts` is the
definition. Each shape is a must-stay-silent fixture.

## The 13 at HEAD

`ReplicateApi.ts:228`, `TasksWatchCommandLineInterface.ts:306`, `IntelligenceMediaGenerationApi.ts:167`,
`GeminiApi.ts:252`, `FigmaMcpLauncher.ts:16`, `RainbowMatrix.ts:1187`, `AhraOsWatchers.ts:1278`,
`AhraOsMonitors.ts:253`, `:572`, `PhiSocialCommandLineInterface.ts:261`, `:286`,
`PhiSocialUpload.ts:286`, and `nexus/source/coordination/testing/NextEventLoopTurn.ts:78`, which holds
the promise in a variable to race it later and is still `delay(turnDurationInMilliseconds)`. The
review's list named six of these; the other seven are the same shape. The two left in the working
tree are `PhiSocialUpload.ts:286` and `NextEventLoopTurn.ts:78`.

## Divergence from ESLint

None to record in `internal/differential/acknowledged.go`: the gate has no rule of this name.
`no-promise-executor-return` stays on; it catches a different bug (an executor that returns a value)
and reported only `ReplicateApi.ts:228` of these, the one with an expression body.

## How the measurement was taken

As for `consistency_no_for_in.md`: real binary on the working tree with a re-anchored scratch config
(3,745 files, 2 findings), and HEAD by parsing a `git archive` of ahra and both submodules.

## Not auto-fixable

The rewrite needs an import whose path depends on the file's package (Nexus imports `delay`
relatively), and the replacement resolves `void` where the original may be typed or used as a value.
