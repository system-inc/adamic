# `guard-for-in`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **4** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require `for-in` loops to include an `if` statement

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

4 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Object.ts:79`**

```
for(const key in updates) {
```

> The body of a for-in should be wrapped in an if statement to filter unwanted properties from the prototype

**`libraries/structure/libraries/nexus/source/collections/Object.ts:112`**

```
for(const currentKey in object) {
```

> The body of a for-in should be wrapped in an if statement to filter unwanted properties from the prototype

**`libraries/structure/libraries/nexus/source/geography/Countries.ts:44`**

```
for(const countryCode in Countries) {
```

> The body of a for-in should be wrapped in an if statement to filter unwanted properties from the prototype

**`libraries/structure/source/theme/utilities/ThemeUtilities.tsx:63`**

```
for(const key in overrideTheme) {
```

> The body of a for-in should be wrapped in an if statement to filter unwanted properties from the prototype

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

