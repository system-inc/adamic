# `no-eq-null`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **57** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow `null` comparisons without type-checking operators

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

57 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/inventory/FinanceInventoryAccountGroup.tsx:47`**

```
const hasPortfolio = financeAccountPortfolio.data?.portfolio != null;
```

> Use '===' to compare with null

**`libraries/structure/libraries/nexus/code-quality/lint/rules/ConsistencyRequireConstantCasingRule.ts:357`**

```
if(annotation.typeArguments != null) return false;
```

> Use '===' to compare with null

**`libraries/structure/libraries/nexus/source/structured-text/json/JsonValueTransformers.ts:61`**

```
if(jsonValue == null) {
```

> Use '===' to compare with null

**`libraries/structure/libraries/nexus/source/structured-text/json/JsonValueTransformers.ts:74`**

```
if(objectValue == null) {
```

> Use '===' to compare with null

**`libraries/structure/libraries/nexus/source/validation/SafeParse.ts:10`**

```
if(value == null) {
```

> Use '===' to compare with null

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

