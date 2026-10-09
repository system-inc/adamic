# `react/no-array-index-key`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **84** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow usage of Array index in keys

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

84 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:45`**

```
key={navigationLinkIndex}
```

> Do not use Array index in keys

**`app/(os-layout)/_components/kingdom/OsKingdomActivityFeed.tsx:30`**

```
key={`${event.occurredAtInMilliseconds}-${index}`}
```

> Do not use Array index in keys

**`app/(os-layout)/_components/kingdom/activity/OsKingdomActivityModelTurn.tsx:21`**

```
<p key={index} className="whitespace-pre-wrap content--1">
```

> Do not use Array index in keys

**`app/(os-layout)/_components/kingdom/activity/OsKingdomActivityModelTurn.tsx:26`**

```
return <OsKingdomActivityToolUse key={index} part={part} />;
```

> Do not use Array index in keys

**`app/(os-layout)/finance/_components/FinanceChart.tsx:192`**

```
<Cell key={pointIndex} fill={point.barColor ?? seriesColor(series, index)} />
```

> Do not use Array index in keys

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

