# `@typescript-eslint/require-array-sort-compare`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **4** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Require `Array#sort` and `Array#toSorted` calls to always provide a `compareFunction`

## Why this recommendation

Already disabled in VerifySettings.json, and 3 of 4 sites are `results.sort()` on small number arrays inside PromiseGroup.test.ts assertions where the default lexicographic sort is harmless.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong Yes** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

4 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/coordination/PromiseGroup.test.ts:88`**

```
expect(results.sort()).toEqual([1, 2, 3]);
```

> Require 'compare' argument

**`libraries/structure/libraries/nexus/source/coordination/PromiseGroup.test.ts:305`**

```
expect(results.sort()).toEqual([1, 2, 3, 4, 5]);
```

> Require 'compare' argument

**`libraries/structure/libraries/nexus/source/coordination/PromiseGroup.test.ts:380`**

```
expect([result1, result2].sort()).toEqual([1, 2]);
```

> Require 'compare' argument

**`modules/tasks/TasksViewsCommandLineInterface.ts:850`**

```
for(const [status, count] of Object.entries(byStatus).sort()) {
```

> Require 'compare' argument

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

