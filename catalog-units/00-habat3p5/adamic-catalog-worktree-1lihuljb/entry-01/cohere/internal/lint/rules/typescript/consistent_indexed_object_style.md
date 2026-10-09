# `@typescript-eslint/consistent-indexed-object-style`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **43** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require or disallow the `Record` type

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

43 in the tree. Showing the first few.

**`libraries/structure/StructureSettings.ts:169`**

```
layers?: {
```

> A record is preferred over an index signature

**`libraries/structure/StructureSettings.ts:182`**

```
platforms: {
```

> A record is preferred over an index signature

**`libraries/structure/libraries/nexus/source/security/random/NonceGeneration.test.ts:64`**

```
const characterCounts: { [key: string]: number } = {};
```

> A record is preferred over an index signature

**`libraries/structure/libraries/nexus/source/types/ObjectTypes.ts:24`**

```
export type AddPropertyType<T, K extends string, V> = T & {
```

> A record is preferred over an index signature

**`libraries/structure/libraries/nexus/source/types/ObjectTypes.ts:29`**

```
export type AddOptionalPropertyType<T, K extends string, V> = T & {
```

> A record is preferred over an index signature

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

