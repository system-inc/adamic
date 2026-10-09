# `@typescript-eslint/return-await`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **32** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce consistent awaiting of returned promises

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

32 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.ts:6`**

```
return await crypto.subtle.generateKey(
```

> Returning an awaited promise is not allowed in this context

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.ts:18`**

```
return await crypto.subtle.generateKey(
```

> Returning an awaited promise is not allowed in this context

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.ts:30`**

```
return await crypto.subtle.deriveKey(
```

> Returning an awaited promise is not allowed in this context

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.ts:61`**

```
return await crypto.subtle.importKey(
```

> Returning an awaited promise is not allowed in this context

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.ts:76`**

```
return await crypto.subtle.importKey(
```

> Returning an awaited promise is not allowed in this context

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

