# `nexus/correctness-no-uncleared-race-timeout`

| | |
|---|---|
| **Recommendation** | **Yes, at `error` once the one ahra site is fixed.** 1 finding, 1 true |
| Findings | **ahra 1** (measured 2026-10-03) |
| Measured precision | 1 of 1 true: the timer's handle is dropped where it is made |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `Promise.race`, `Promise` and `setTimeout` to the default library and the global declarations; without a checker the rule registers nothing |

## What it checks

A `setTimeout` armed directly in the executor of a `new Promise(...)` that is an element of a
`Promise.race([...])` (in place, or through a `const` of the same file), whose handle nothing can reach:
a statement on its own, the executor's expression body, under `void`, or stored in a binding nothing in
the file reads. When the other side wins, that timer stays armed for its whole duration. The doc comment
carries the precise shape and what stays silent.

## Where it came from

`modules/phi/social/PhiSocialGenerator.ts:578` in ahra: the per-image render bound races
`generateImage(...)` against a `new Promise<never>` whose executor arms a ten-minute `setTimeout` and keeps
nothing, so an unattended batch CLI lingers up to ten minutes after its last image. Found by the
cross-language pass of the new-rules sweep (`#tevhg3f`, after go vet `lostcancel`), built in `#j03vwm6`.

## Reconciling with the research count

Research counted **1** (`V4_promise_race_timer_leak`). cohere counts **1**, the same site. The task
believed this site fixed; it is still live at HEAD on 2026-10-03, so the count is 1, not 0. Every
`Promise.race` in ahra was read: the other two (`TrackedPromise.test.ts:290`, `NeverResolves.test.ts:22`)
race promises with no timer.

## What it declines

A timeout promise from a helper (`timeoutAfter(ms)`), one reaching the array through a spread, a `let`
or a property, and a timer armed in a callback nested inside the executor. Each is a missed finding.
Every use of a kept handle other than a plain `=` write counts as a possible clear, so a handle that is
logged and never cleared is missed too.

## Verification

- Fixtures from the real site both ways: `PhiSocialGenerator`'s race as it stands (one finding, on the
  `setTimeout`) and fixed (`imageTimer = setTimeout(...)`, `clearTimeout(imageTimer)` in a `finally`,
  silent). 7 firing shapes, 11 silent shapes, the `@types/node` `declare global` declaration (with the
  merged `namespace setTimeout`) and lib.dom's (`window.setTimeout`, which resolves to the global function
  and the `WindowOrWorkerGlobalScope` member at once).
- Mutation check, each mutant run through `go test -overlay -count=1`, all killed: handle loss not
  required (handle cleared in a finally, `.unref()`), any `race` method (an object's own `race`), any
  `setTimeout` (a local function of that name), binding reads never seen (a const handle cleared later),
  plain assignments counted as reads (a handle assigned to an outer binding nothing reads).
