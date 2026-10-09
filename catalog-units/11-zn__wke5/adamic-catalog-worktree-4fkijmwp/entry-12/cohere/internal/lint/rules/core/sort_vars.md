# `sort-vars`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **2** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require variables within the same declaration block to be sorted

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

2 in the tree. Showing the first few.

**`libraries/structure/source/components/charts/time-series/hooks/useZoomBehavior.tsx:60`**

```
let startIntervalTime: Date, endIntervalTime: Date;
```

> Variables within the same declaration block should be sorted alphabetically

**`modules/google/analytics/AnalyticsApi.ts:672`**

```
let currentFormat: string, previousFormat, changeFormat: string;
```

> Variables within the same declaration block should be sorted alphabetically

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

