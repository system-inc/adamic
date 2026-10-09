# `@typescript-eslint/prefer-enum-initializers`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **3** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require each enum member value to be explicitly initialized

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

3 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:386`**

```
FIRST, // 0
```

> The value of the member 'FIRST' should be explicitly defined

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:387`**

```
SECOND, // 1
```

> The value of the member 'SECOND' should be explicitly defined

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:388`**

```
THIRD, // 2
```

> The value of the member 'THIRD' should be explicitly defined

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

