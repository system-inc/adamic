# `@typescript-eslint/no-mixed-enums`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **1** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow enums from having both number and string members

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

1 in the tree. 

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:415`**

```
NUMERIC_VALUE = 42,
```

> Mixing number and string enums can be confusing

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

