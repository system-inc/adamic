# `radix`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **172** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce the use of the radix argument when using `parseInt()`

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

172 in the tree. Showing the first few.

**`app/api/conversations/route.ts:11`**

```
const limit = parseInt(url.searchParams.get('limit') ?? '30');
```

> Missing radix parameter

**`app/api/tasks/route.ts:62`**

```
const limit = parseInt(url.searchParams.get('limit') ?? '50');
```

> Missing radix parameter

**`libraries/structure/libraries/nexus/source/encoding/StreamTransforms.ts:84`**

```
const dataLength = parseInt(prefixString);
```

> Missing radix parameter

**`libraries/structure/libraries/nexus/source/security/cryptography/DecryptionOperations.ts:85`**

```
const type = parseInt(kind) as EncryptedDataKindType;
```

> Missing radix parameter

**`libraries/structure/libraries/nexus/source/time/Duration.ts:152`**

```
const hours = parseInt(match[1] || '0');
```

> Missing radix parameter

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

