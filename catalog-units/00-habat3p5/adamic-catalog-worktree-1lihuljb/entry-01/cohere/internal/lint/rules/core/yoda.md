# `yoda`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **7** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require or disallow "Yoda" conditions

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

7 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/errors/Assert.test.ts:16`**

```
expect(() => assert('hello' === ('world' as string))).toThrow('Assertion failed');
```

> Expected literal to be on the right side of ===

**`libraries/structure/source/ops/design/colors/components/OpsDesignColorsScaleGenerator.tsx:317`**

```
if(0 <= hue && hue < 60) {
```

> Expected literal to be on the right side of <=

**`libraries/structure/source/ops/design/colors/components/OpsDesignColorsScaleGenerator.tsx:322`**

```
else if(60 <= hue && hue < 120) {
```

> Expected literal to be on the right side of <=

**`libraries/structure/source/ops/design/colors/components/OpsDesignColorsScaleGenerator.tsx:327`**

```
else if(120 <= hue && hue < 180) {
```

> Expected literal to be on the right side of <=

**`libraries/structure/source/ops/design/colors/components/OpsDesignColorsScaleGenerator.tsx:332`**

```
else if(180 <= hue && hue < 240) {
```

> Expected literal to be on the right side of <=

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

