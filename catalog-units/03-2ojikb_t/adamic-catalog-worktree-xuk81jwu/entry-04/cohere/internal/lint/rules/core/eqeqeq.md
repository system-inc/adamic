# `eqeqeq`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **75** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require the use of `===` and `!==`

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

75 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/inventory/FinanceInventoryAccountGroup.tsx:47`**

```
const hasPortfolio = financeAccountPortfolio.data?.portfolio != null;
```

> Expected '!==' and instead saw '!='

**`libraries/structure/libraries/nexus/code-quality/lint/rules/ConsistencyRequireConstantCasingRule.ts:357`**

```
if(annotation.typeArguments != null) return false;
```

> Expected '!==' and instead saw '!='

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyOperations.ts:12`**

```
if(typeof data == 'string') data = new TextEncoder().encode(data);
```

> Expected '===' and instead saw '=='

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyOperations.ts:24`**

```
if(typeof data == 'string') data = new TextEncoder().encode(data);
```

> Expected '===' and instead saw '=='

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyOperations.ts:30`**

```
if(typeof data == 'string') {
```

> Expected '===' and instead saw '=='

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

