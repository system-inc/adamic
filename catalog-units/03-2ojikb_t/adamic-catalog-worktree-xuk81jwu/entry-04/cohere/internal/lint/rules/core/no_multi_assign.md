# `no-multi-assign`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **3** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow use of chained assignment expressions

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

3 in the tree. Showing the first few.

**`baselines/DumpRuleInventory.ts:159`**

```
const bucket = (byPlugin[record.plugin] ??= { known: 0, enabled: 0 });
```

> Unexpected chained assignment

**`libraries/structure/source/components/code/Code.tsx:67`**

```
target.selectionStart = target.selectionEnd = start + 2;
```

> Unexpected chained assignment

**`modules/finance/connections/FinanceQuickBooksEnricher.ts:261`**

```
const bucket = (samples[plan.outcome] ??= []);
```

> Unexpected chained assignment

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

