# `no-param-reassign`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **42** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow reassigning function parameters

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

42 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/encoding/DynamicBuffer.ts:81`**

```
newSize = Math.max(this.buffer.byteLength * 2, newSize);
```

> Assignment to function parameter 'newSize'

**`libraries/structure/libraries/nexus/source/errors/BaseErrorSerializer.ts:124`**

```
seen = new WeakSet();
```

> Assignment to function parameter 'seen'

**`libraries/structure/libraries/nexus/source/protocols/http/Cookies.ts:15`**

```
domain = domain.trim();
```

> Assignment to function parameter 'domain'

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyOperations.ts:12`**

```
if(typeof data == 'string') data = new TextEncoder().encode(data);
```

> Assignment to function parameter 'data'

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyOperations.ts:24`**

```
if(typeof data == 'string') data = new TextEncoder().encode(data);
```

> Assignment to function parameter 'data'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

