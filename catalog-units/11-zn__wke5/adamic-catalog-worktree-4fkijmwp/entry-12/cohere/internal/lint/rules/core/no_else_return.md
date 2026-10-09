# `no-else-return`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **31** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow `else` blocks after `return` statements in `if` statements

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

31 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/media/images/Image.ts:200`**

```
else {
```

> Unnecessary 'else' after 'return'

**`libraries/structure/libraries/nexus/source/security/cryptography/DecryptionOperations.ts:93`**

```
else {
```

> Unnecessary 'else' after 'return'

**`libraries/structure/libraries/nexus/source/structured-text/json/JsonValueTransformers.ts:41`**

```
else {
```

> Unnecessary 'else' after 'return'

**`libraries/structure/libraries/nexus/source/time/FormatTime.ts:169`**

```
else {
```

> Unnecessary 'else' after 'return'

**`libraries/structure/libraries/nexus/source/time/FormatTime.ts:323`**

```
else {
```

> Unnecessary 'else' after 'return'

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

