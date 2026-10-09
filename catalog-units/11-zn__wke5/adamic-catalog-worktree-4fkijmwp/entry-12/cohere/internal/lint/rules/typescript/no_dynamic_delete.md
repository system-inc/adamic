# `@typescript-eslint/no-dynamic-delete`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **16** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow using the `delete` operator on computed key expressions

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

16 in the tree. Showing the first few.

**`libraries/structure/command-line/Structure.ts:220`**

```
delete environment[verifyBinaryOverrideVariable];
```

> Do not delete dynamically computed property keys

**`libraries/structure/libraries/nexus/source/protocols/http/CloudflareRequest.ts:20`**

```
delete out[key];
```

> Do not delete dynamically computed property keys

**`libraries/structure/source/components/tables/hooks/useTable.ts:478`**

```
else delete copy[rowId];
```

> Do not delete dynamically computed property keys

**`libraries/structure/source/components/tables/hooks/useTable.ts:500`**

```
else delete next[row.id];
```

> Do not delete dynamically computed property keys

**`libraries/structure/source/components/tables/hooks/useTable.ts:528`**

```
delete next[columnId];
```

> Do not delete dynamically computed property keys

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

