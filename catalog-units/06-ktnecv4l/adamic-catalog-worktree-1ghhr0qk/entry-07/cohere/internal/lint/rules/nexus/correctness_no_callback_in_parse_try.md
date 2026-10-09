# `nexus/correctness-no-callback-in-parse-try`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`.** ahra is clean today, so it can be enabled with nothing to fix; it exists to stop the shape coming back |
| Findings | **ahra 0** (measured 2026-10-03). **2 on ahra at `1382cbe9^`**, the tree the research probe read, matching its count exactly |
| Measured precision | 2 of 2 true on the pre-fix tree, both the real sites below |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `JSON.parse`, nexus `parseJsonOrThrow`, parameters and `new Promise`; without a checker the rule registers nothing |

## What it checks

A call to a caller-supplied function inside a `try` whose block also parses JSON, when the `catch`
throws the caught error away. The callback's own error is then caught by a handler written for the
parse, so it is handled as malformed input and the real error is lost.

- **The parse**: `JSON.parse` (the method on the global `JSON` interface, by declaration) or nexus
  `parseJsonOrThrow` (the function declared in `source/structured-text/json/Json.ts`, through any import
  alias). `parseJson` returns an outcome and never throws, so it does not count.
- **The callback**: a call whose callee is a bare name bound by a parameter, directly or destructured
  out of one, whose non-nullish type has a call signature. Methods on a parameter (`logger.info`) and
  `any` parameters are not counted. The resolve and reject of a `new Promise` executor (the global
  Promise) are exempt: they never throw.
- **The catch**: has no binding, or a plain binding it never reads.
- **Where**: both calls in code the `catch` guards, so not inside a nested function and not inside the
  block of a nested `try` that has its own `catch` (that nested `catch` and its `finally` are walked).

The callback call is reported, once per call.

## Where it came from

The own-history pass of the new-rules sweep (`#tevhg3f`, probe P5), built as `#hspse4v`:

- `modules/claude/utilities/ClaudeUtilities.ts`, `parseStreamLines` (`#wv045mc`): `try { const message =
  JSON.parse(line); onMessage(message); } catch { onUnparseable?.(line); }`. A handler that threw on a
  well-formed line was reported as an unparseable line.
- `modules/x/XStreamApi.ts`, `consumeStream`: `try { const parsed = JSON.parse(line) as ...;
  onEvent(parsed); } catch { // Heartbeat or partial frame; skip. }`. Every handler error vanished.

Both were fixed in the JSON waves (`1382cbe9`, `4e385b61`) by moving to `parseJson` and calling the
callback on a `'Parsed'` outcome, outside any `try`.

## The design decision: what "a catch that reports a parse error" means

The brief's precision line asks for a `catch` that *reports a parse error*. Whether a `catch` says
"parse error" lives in its message text or its comment, and reading those is a name heuristic. So the
rule reads the one thing the AST says exactly: whether the `catch` can tell the callback's error from
the parse error at all. A `catch` that never reads its caught value treats every error as the failure it
was written for, which, around a parse, is the parse.

Declined, on purpose: a `catch` that reads its binding. That covers the correct shapes (rethrowing it,
`if(!(error instanceof SyntaxError)) throw error`, handing it to an `onError` channel as itself) and
also some that still mislabel it (`log('bad line', error)`, `onUnparseable(line, error)`). Telling those
apart means reading what the receiver does with the error, which is a guess. A mislabelling `catch` that
reads its binding is a missed finding, never a false one.

**What that costs, measured**: a variant with the `catch` condition removed entirely (any `catch` at all)
finds the same 2 on the pre-fix tree and the same 0 on today's ahra. On ahra the narrowing loses nothing.

## Reconciling with the research count

Research counted **2** (probe P5, on 2026-10-01 before the JSON waves). cohere counts **0** on today's
ahra and **2** on the tree at `1382cbe9^` (the parent of the first of the two fix commits, ahra plus
the Structure submodule at its pinned commit, extracted read-only with `git archive`):
`ClaudeUtilities.ts:228` (`onMessage(message)`) and `XStreamApi.ts:47` (`onEvent(parsed)`; research
named `:45`, the `try` line). The gap is the two fixes, nothing else.

## Verification

- Fixtures both ways from the real sites: `parseStreamLines` and `consumeStream` before (one finding
  each, on the handler call and not on the fallback in the `catch`) and after (the `parseJson` form,
  silent). 13 firing shapes and 20 silent shapes besides, including a module-scoped `JSON` interface,
  `JSON.stringify`, a lookalike `parseJsonOrThrow`, a local `Promise` class, closures and nested tries.
- Mutation check, each mutant a copy through `go test -overlay -count=1`, each killed: the `catch`
  condition (four silent fixtures fire), the parse requirement, the Promise executor exemption, the
  global scope test for `Promise` (local class) and for `JSON` (module interface), the `parse` name on
  `JSON` (`JSON.stringify`), the nexus file suffix (lookalike), alias resolution, the `parseJsonOrThrow`
  match, the nested-function boundary, the nested `try` skip, walking a nested `catch`, the parameter
  test (local function), destructured parameters, the non-nullable call signature (`?.`, `!`),
  unwrapping `!` and parentheses, shorthand reads of the binding. **Survived**: dropping the explicit
  "binding is a plain name" test on the `catch`, because a destructured binding has no symbol and the
  next check refuses it too; it stays as the statement of intent.
