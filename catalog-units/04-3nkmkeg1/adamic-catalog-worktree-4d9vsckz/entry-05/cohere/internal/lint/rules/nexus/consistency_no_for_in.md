# `nexus/consistency-no-for-in`

| | |
|---|---|
| **Recommendation** | **Yes**, in place of `guard-for-in` (Kirk's ruling, 2026-10-01: one loop shape for objects) |
| Findings in ahra | **8** at HEAD on 2026-10-01; **0** in the working tree after the syntax wave rewrote them |
| Measured precision | 8 of 8 by the rule's definition (every `for...in` is the shape being retired) |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Every `for...in` statement. Write `for (const [key, value] of Object.entries(object))`, or
`for (const key of Object.keys(object))` when only the key matters. The `in` operator, `for...of`
and `for await...of` are not reported.

## Why it replaces `guard-for-in`

`guard-for-in` asks each `for...in` to open with an `if`, and cannot tell a guard from any other
`if`: `ReactComponentNoDestructuringRule.ts:80` opens with `if (key === 'parent') continue;`, passes
upstream, and still visits inherited keys. `Object.entries` and `Object.keys` return own enumerable
string keys, which is what a correctly guarded loop computes, so the replacement has the guard built
in, and it removes the `key as keyof typeof object` casts these loops carry to read the value.

## Divergence from ESLint

None to record in `internal/differential/acknowledged.go`: the gate has no rule of this name, so
there is no shared rule for the two to disagree on. It is stricter than `guard-for-in` by design
(it also reports the four loops `guard-for-in` accepts as guarded), which the parity doctrine
allows, and `guard-for-in` should be turned off when this is enabled so one loop is not judged by
two rules that ask for different repairs.

## The 8 at HEAD

Reported by `guard-for-in` too: `nexus/source/geography/Countries.ts:883`,
`source/theme/utilities/ThemeUtilities.tsx:65`, `nexus/source/collections/Object.ts:79`, `:112`.
Accepted by `guard-for-in` as guarded: `Object.ts:44` (`Object.hasOwn`),
`useTableRowSelectionSubscription.ts:73` and `ClassName.ts:29` (a single `if`),
`code-quality/lint/rules/ReactComponentNoDestructuringRule.ts:80` (a `continue` that is not a guard).
All eight are fixtures in `consistency_no_for_in_test.go`.

## How the measurement was taken

Binary built from this tree into a scratchpad and run from `~/Projects/ahra` as
`cohere --no-fix --lint --lint-config <scratch copy>` with the three new nexus rules enabled and the
copy's non-`**` globs re-anchored to ahra (a config's globs resolve against its own directory):
3,745 files, 0 findings for this rule in the working tree. HEAD was measured by parsing a
`git archive` of ahra and both submodules with the rule directly (the rule reads no types).

## Not auto-fixable

Choosing `Object.keys` or `Object.entries`, renaming the value reads in the body, and the rare loop
whose object has an enumerable prototype property (where `Object.keys` visits fewer keys) are the
author's call.
