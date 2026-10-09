# `no-bitwise`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **82** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow bitwise operators

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

82 in the tree. Showing the first few.

**`app/(os-layout)/os/providers/OsProvidersPalette.ts:53`**

```
hash = ((hash * 33) ^ profileId.charCodeAt(index)) >>> 0;
```

> Unexpected use of '>>>'

**`app/(os-layout)/os/providers/OsProvidersPalette.ts:53`**

```
hash = ((hash * 33) ^ profileId.charCodeAt(index)) >>> 0;
```

> Unexpected use of '^'

**`libraries/structure/libraries/nexus/code-quality/lint/utilities/BrandTypeUtilities.ts:63`**

```
const nonUndefined = brandType.types.filter((member) => (member.flags & TypeScript.TypeFlags.Undefined) === 0);
```

> Unexpected use of '&'

**`libraries/structure/libraries/nexus/code-quality/lint/utilities/DecoratorTypeUtilities.ts:40`**

```
TypeScript.TypeFlags.Null |
```

> Unexpected use of '|'

**`libraries/structure/libraries/nexus/code-quality/lint/utilities/DecoratorTypeUtilities.ts:40`**

```
TypeScript.TypeFlags.Null |
```

> Unexpected use of '|'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

