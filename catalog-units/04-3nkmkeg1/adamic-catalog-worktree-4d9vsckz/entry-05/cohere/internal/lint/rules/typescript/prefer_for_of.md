# `@typescript-eslint/prefer-for-of`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **5** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce the use of `for-of` loop over the standard `for` loop where possible

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

5 in the tree. Showing the first few.

**`libraries/structure/source/components/charts/time-series/TimeSeriesBar.tsx:34`**

```
for(let index = 0; index < stackedDataKeys.length; index++) {
```

> Expected a `for-of` loop instead of a `for` loop with this simple iteration

**`libraries/structure/source/components/tables/columns/hooks/useTableColumnFitToContent.ts:148`**

```
for(let propertyIndex = 0; propertyIndex < inlineStyle.length; propertyIndex++) {
```

> Expected a `for-of` loop instead of a `for` loop with this simple iteration

**`libraries/structure/source/utilities/style/ClassName.ts:17`**

```
for(let index = 0; index < value.length; index++) {
```

> Expected a `for-of` loop instead of a `for` loop with this simple iteration

**`libraries/structure/source/utilities/style/ClassName.ts:44`**

```
for(let index = 0; index < values.length; index++) {
```

> Expected a `for-of` loop instead of a `for` loop with this simple iteration

**`modules/google/docs/DocsApi.ts:374`**

```
for(let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
```

> Expected a `for-of` loop instead of a `for` loop with this simple iteration

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

