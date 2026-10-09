# `@typescript-eslint/prefer-reduce-type-parameter`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **2** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce using type parameter when calling `Array#reduce` instead of using a type assertion

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

2 in the tree. Showing the first few.

**`libraries/structure/source/localization/Translations.ts:69`**

```
{} as Record<string, unknown>,
```

> Unnecessary assertion: Array#reduce accepts a type parameter for the default value

**`libraries/structure/source/utilities/style/ClassName.ts:364`**

```
{} as Record<string, unknown>,
```

> Unnecessary assertion: Array#reduce accepts a type parameter for the default value

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

