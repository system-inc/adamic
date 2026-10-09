# `@typescript-eslint/no-duplicate-type-constituents`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **3** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow duplicate constituents of union or intersection types

## Why this recommendation

Only 3 sites and one is in generated code (GraphQlOperations.ts), so it is autofixable housekeeping rather than a defect worth top priority.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong Yes** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

3 in the tree. Showing the first few.

**`libraries/structure/source/api/graphql/generated/GraphQlOperations.ts:1822`**

```
constructor(value: string, __meta__?: Record<string, unknown> | undefined) {
```

> Explicit undefined is unnecessary on an optional parameter

**`libraries/structure/source/ops/developers/metrics/DataSource.tsx:46`**

```
options?: NextUseQueryStateOptions | undefined,
```

> Explicit undefined is unnecessary on an optional parameter

**`libraries/structure/source/ops/developers/metrics/DataSources.tsx:38`**

```
options?: NextUseQueryStateOptions | undefined,
```

> Explicit undefined is unnecessary on an optional parameter

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

