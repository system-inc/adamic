# `@typescript-eslint/prefer-literal-enum-member`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **2** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require all enum members to be literal values

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

2 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:429`**

```
B = A * 2,
```

> Explicit enum value must only be a literal value (string or number)

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:430`**

```
C = B + A,
```

> Explicit enum value must only be a literal value (string or number)

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

