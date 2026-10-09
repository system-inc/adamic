# `@typescript-eslint/no-unnecessary-qualifier`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **1** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow unnecessary namespace qualifiers

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

1 in the tree. 

**`libraries/structure/libraries/nexus/source/protocols/http/Cookies.ts:99`**

```
const domain = Cookies.normalizeDomain(options.domain);
```

> Qualifier is unnecessary since 'normalizeDomain' is in scope

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

