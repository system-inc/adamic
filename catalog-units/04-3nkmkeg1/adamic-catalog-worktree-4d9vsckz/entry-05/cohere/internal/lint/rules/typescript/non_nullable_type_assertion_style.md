# `@typescript-eslint/non-nullable-type-assertion-style`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **55** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce non-null assertions over explicit type assertions

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

55 in the tree. Showing the first few.

**`app/(os-layout)/os/wisdom/_components/WisdomLeverageCard.tsx:131`**

```
cy={pointY(properties.values[lastIndex] as number)}
```

> Use a ! assertion to more succinctly remove null and undefined from the type

**`libraries/structure/libraries/nexus/source/time/cron/CronExpression.ts:279`**

```
const rangePart = slashParts[0] as string;
```

> Use a ! assertion to more succinctly remove null and undefined from the type

**`libraries/structure/libraries/nexus/source/time/cron/CronExpression.ts:299`**

```
if(!/^\d+$/.test(dashParts[0] as string) || !/^\d+$/.test(dashParts[1] as string)) {
```

> Use a ! assertion to more succinctly remove null and undefined from the type

**`libraries/structure/libraries/nexus/source/time/cron/CronExpression.ts:299`**

```
if(!/^\d+$/.test(dashParts[0] as string) || !/^\d+$/.test(dashParts[1] as string)) {
```

> Use a ! assertion to more succinctly remove null and undefined from the type

**`libraries/structure/libraries/nexus/source/time/cron/CronExpression.ts:651`**

```
if(matching.length > 0) return matching[0] as number;
```

> Use a ! assertion to more succinctly remove null and undefined from the type

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

