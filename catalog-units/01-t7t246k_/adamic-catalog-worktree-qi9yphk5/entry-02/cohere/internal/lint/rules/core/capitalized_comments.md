# `capitalized-comments`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **15110** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce or disallow capitalization of the first letter of a comment

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

15110 in the tree. Showing the first few.

**`LintConfiguration.ts:69`**

```
// with a `fired` flag: the ruling banks even when the
```

> Comments should not begin with a lowercase character

**`LintConfiguration.ts:70`**

```
// executor misses, and web needs both bits to say
```

> Comments should not begin with a lowercase character

**`ProjectRoot.ts:40`**

```
// root's package.json and requires this exact value, so a path that merely looks
```

> Comments should not begin with a lowercase character

**`ProjectRoot.ts:41`**

```
// plausible (a parent directory, another checkout, a nested copy) is rejected loudly
```

> Comments should not begin with a lowercase character

**`ProjectRoot.ts:42`**

```
// instead of being silently populated with a duplicate data tree.
```

> Comments should not begin with a lowercase character

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

