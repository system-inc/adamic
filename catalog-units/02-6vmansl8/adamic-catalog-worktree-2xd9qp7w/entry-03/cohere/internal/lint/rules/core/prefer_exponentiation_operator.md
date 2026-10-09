# `prefer-exponentiation-operator`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **29** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow the use of `Math.pow` in favor of the `**` operator

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

29 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/coordination/Delay.ts:44`**

```
let backoff = Math.min(minimumMilliseconds * Math.pow(2, attempt - 1), maximumMilliseconds);
```

> Use the '**' operator instead of 'Math.pow'

**`libraries/structure/libraries/nexus/source/files/File.ts:168`**

```
const multiplier = Math.pow(10, decimals);
```

> Use the '**' operator instead of 'Math.pow'

**`libraries/structure/libraries/nexus/source/files/File.ts:243`**

```
const value = bytes / Math.pow(base, forcedUnitIndex);
```

> Use the '**' operator instead of 'Math.pow'

**`libraries/structure/libraries/nexus/source/files/File.ts:255`**

```
const value = bytes / Math.pow(base, unitIndex);
```

> Use the '**' operator instead of 'Math.pow'

**`libraries/structure/libraries/nexus/source/numbers/Statistics.ts:147`**

```
return accumulator + Math.pow(current - average, 2);
```

> Use the '**' operator instead of 'Math.pow'

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

