# `prefer-promise-reject-errors`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **1** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require using Error objects as Promise rejection reasons

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

1 in the tree. 

**`modules/kingdom/pentair/PentairApi.ts:78`**

```
reject(error as Error);
```

> Expected the Promise rejection reason to be an Error

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

