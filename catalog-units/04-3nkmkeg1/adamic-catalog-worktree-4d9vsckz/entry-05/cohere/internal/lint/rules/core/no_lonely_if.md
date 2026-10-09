# `no-lonely-if`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **16** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow `if` statements as the only statement in `else` blocks

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

16 in the tree. Showing the first few.

**`libraries/structure/source/components/charts/time-series/utilities/TimeSeriesFormatters.tsx:318`**

```
if(quarter === 'Q1') {
```

> Unexpected if as the only statement in an else block

**`libraries/structure/source/components/charts/time-series/utilities/TimeSeriesFormatters.tsx:350`**

```
if(currentMonth === '01') {
```

> Unexpected if as the only statement in an else block

**`libraries/structure/source/components/charts/time-series/utilities/TimeSeriesFormatters.tsx:407`**

```
if(day === 1 || day === 15) {
```

> Unexpected if as the only statement in an else block

**`libraries/structure/source/components/charts/time-series/utilities/TimeSeriesFormatters.tsx:560`**

```
if(hour % 3 === 0) {
```

> Unexpected if as the only statement in an else block

**`libraries/structure/source/components/charts/time-series/utilities/TimeSeriesFormatters.tsx:584`**

```
if(day % 5 === 0 || day === 1) {
```

> Unexpected if as the only statement in an else block

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

