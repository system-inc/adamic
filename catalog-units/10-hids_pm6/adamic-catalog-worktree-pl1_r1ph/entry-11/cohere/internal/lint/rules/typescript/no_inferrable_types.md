# `@typescript-eslint/no-inferrable-types`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **203** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow explicit type declarations for variables or parameters initialized to a number, string, or boolean

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

203 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Array.ts:69`**

```
function getRandom<T>(array: ReadonlyArray<T>, strict: boolean = true): T | undefined {
```

> Type boolean trivially inferred from a boolean literal, remove type annotation

**`libraries/structure/libraries/nexus/source/coordination/BackoffTask.ts:34`**

```
private isRunning: boolean = false;
```

> Type boolean trivially inferred from a boolean literal, remove type annotation

**`libraries/structure/libraries/nexus/source/coordination/BackoffTask.ts:35`**

```
private isInterrupted: boolean = false;
```

> Type boolean trivially inferred from a boolean literal, remove type annotation

**`libraries/structure/libraries/nexus/source/coordination/BackoffTask.ts:36`**

```
private attemptCount: number = 0;
```

> Type number trivially inferred from a number literal, remove type annotation

**`libraries/structure/libraries/nexus/source/coordination/Delay.ts:34`**

```
jitter: boolean = true,
```

> Type boolean trivially inferred from a boolean literal, remove type annotation

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

## Status: ported 2026-09-06, registered and NOT enabled

The audit's 203 and the port's 204 differ by one, and the cause is the instrument rather than either
count. The audit measured through ESLint with inline `eslint-disable` comments active; cohere applies
suppression outside the rule, so a rule fixture and a rule dry-run both see the pre-suppression
verdict. Driving the installed 8.67.0 rule over the same 89 files with `noInlineConfig` gives 204,
and all 204 agree with cohere position for position, columns included.

Registered so it compiles in and appears in `cohere --rules`; deliberately not enabled, because
enabling it means accepting 204 findings' worth of cleanup, and that is a decision about this
codebase rather than a porting step. The audit's **No** above is a judgment about that cleanup cost
and it is untouched; what has changed is only that the rule now exists to be turned on.

