# `@typescript-eslint/member-ordering`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **852** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require a consistent member declaration order

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

852 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/OsKingdomLayout.ts:74`**

```
[key: string]: unknown; // React Flow types edge data as a Record.
```

> Member key should be declared before all field definitions

**`app/(os-layout)/finance/_components/FinanceChart.tsx:57`**

```
[dataKey: string]: string | number | undefined;
```

> Member dataKey should be declared before all field definitions

**`libraries/structure/StructureSettings.ts:35`**

```
[name: string]: Partial<BaseApiInterface> | undefined;
```

> Member name should be declared before all field definitions

**`libraries/structure/StructureSettings.ts:131`**

```
[key: string]: {
```

> Member key should be declared before all field definitions

**`libraries/structure/StructureSettings.ts:138`**

```
[key: string]: BaseApiInterface | undefined;
```

> Member key should be declared before all field definitions

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

