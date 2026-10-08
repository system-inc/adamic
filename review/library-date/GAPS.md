# Date TS2345 investigation

**Decision superseding the original diagnosis:** stock TypeScript rejects the
coercion tests with the same TS2345. These refusals are correct for Adamic, not
a language gap to fix. The original evidence below is retained as history.
The follow-up audit and runner work are recorded in
[the Date refusal report](../library-date-refusals/report.md).

Observed on `codex/library-date`, compiler commit
`7d532eb9d6699bf8717e1f3ee520c8710a4cba33`, with test262 pinned to
`3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd` and optional `--adapt` disabled.

The 100 refusals are mixed: 83 expose a real object-to-number conversion gap;
17 come from the runner's adapted assertion parameter types. The normalized
TS2345 label collapses these different parameter types into one bucket.
Implementation stops here as requested upon proving the language gap.
Neither the compiler nor the runner's adaptation has been changed.

| First diagnostic | Tests | Origin |
|---|---:|---|
| Object with valueOf or toString passed to a number parameter | 83 | Language/library gap |
| Date passed to assertion primitive union | 12 | Runner harness |
| null passed to assertion primitive union | 4 | Runner harness |
| string or null passed to assertion primitive union | 1 | Runner harness |
| Total | 100 | |

These counts classify each test's **first** diagnostic, not every possible
blocker in its body. Complete paths, unnormalized diagnostics and source context
are recorded in ts2345-diagnostics.json. A temporary audit called the runner's
actual classify(..., false) for every non-skipped Date test, then compiled its
program with the same built adamic command. It found exactly 100 first TS2345
errors. The temporary audit test was removed after recording its output.

## Smallest reproducer matching the coercion tests

File: gap-date-coercion.ts. No harness, var, casts, prototype changes, wall clock,
or local-time operation is involved:

```ts
new Date(0).setUTCDate({ valueOf: (): number => 1 });
```

```
TZ=UTC /tmp/date-ts2345-runner/adamic c review/library-date/gap-date-coercion.ts
```

Observed compiler exit 1:

```
error TS2345: Argument of type '{ valueOf: () => number; }' is not assignable to parameter of type 'number'.
```

Node 24 runs the same source, exit 0, empty stdout and stderr. Observing the
expression's result with Node's console.log prints `0`: ToNumber invokes valueOf,
gets day 1, and returns the epoch timestamp. The numeric control
`new Date(0).setUTCDate(1);` compiles successfully, exit 0.
An even shorter primitive probe `new Date(0).setUTCDate("1");` is also rejected
with TS2345 while Node accepts it, but the object probe matches the 83 tests.

The immediate TS2345 originates in the bundled Date declaration's
`setUTCDate(date: number): number`, not in assertion adaptation. Other setter
and component-constructor parameters are likewise numeric. Widening a harness
assertion signature would not change these call sites.

There is also an actual lowering gap behind that declaration. The second
one-line reproducer, gap-number-coercion.ts, avoids Date's numeric parameter
restriction and reaches the shared conversion path:

```ts
Number({ valueOf: (): number => 1 });
```

Observed compiler exit 1:

```
stage 0 can't lower Number conversion of an object yet
```

Node accepts it, exit 0; observing its result prints `1`.
`internal/lower/library_date.go` routes Date numeric arguments through
libraryNumber; `internal/lower/library_math_number.go` explicitly refuses object
conversion instead of executing valueOf/toString. This proves that merely
widening the Date declarations would leave the required behavior unbuilt.
A full implementation would need the coercion calls, their side effects and
exceptions, and their ordering relative to Date internal-slot reads. Replacing
an object with its eventual numeric value in the runner would change precisely
what these tests exercise.

## Harness contribution

`cmd/adamic-test262/prelude.go` declares assertion operands as
`string | number | boolean | undefined`, excluding object identity and null.
The 12 Date cases use assert.notSameValue on a Date versus undefined; six also
compare Date.prototype. The four null cases are rendered-format tests comparing
a RegExp match with null. The remaining case is toJSON/invoke-result.js, where
the Date library correctly exposes string | null but the assertion excludes null.
These are runner-imposed type restrictions, not evidence that Date arithmetic
is wrong. They remain recorded for a separate runner task after the requested
stop at the real language gap. Removing their first diagnostic is not a promise
that those entire tests would pass: prototype materialization, var, untyped
callbacks and generic borrowed toJSON receivers may still block them.

## Before and after

Both runs used exactly:

```
source /workspace/adamic-tools/env.sh
TZ=UTC go run ./cmd/adamic-test262 -json -work /tmp/date-ts2345-runner -test262 /workspace/test262 built-ins/Date > /tmp/date-ts2345-before.json 2> /tmp/date-ts2345-before.log
TZ=UTC go run ./cmd/adamic-test262 -json -work /tmp/date-ts2345-after-runner -test262 /workspace/test262 built-ins/Date > /tmp/date-ts2345-after.json 2> /tmp/date-ts2345-after.log
```

The runner appends TZ=UTC to compiler, native and Node child environments.

| Measurement | Pass | Disagreements | Refused | Crashes | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| Before | 47 | 0 | 277 | 0 | 270 | 594 |
| After | 47 | 0 | 277 | 0 | 270 | 594 |

TS2345 remains 100. Full reports and directory tables are ts2345-before.json,
ts2345-after.json and their corresponding logs. The figures are unchanged because
this is a diagnosis and stopping point, with review artifacts only.

The diagnostic audit passed in 53.708s. Both complete runner runs exited 0.
Minimal probes, compiler refusals, numeric control and Node observations are
preserved in ts2345-probes.log. No compiler or runtime fix was made, so no new
behavior mutant, oracle fixture or allocation-count row was added. The broad
package gate was not rerun for these review-only files.
