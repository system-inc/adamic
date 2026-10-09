# `@typescript-eslint/no-useless-default-assignment`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **1** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow default values that will never be used

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

1 in the tree. 

**`libraries/structure/libraries/nexus/source/security/random/Random.ts:11`**

```
customizedCharacterSet: string | undefined = undefined,
```

> Using `= undefined` to make a parameter optional adds unnecessary runtime logic. Use the `?` optional syntax instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

