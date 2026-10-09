# `no-continue`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1143** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow `continue` statements

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

1143 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/OsKingdomLayout.ts:385`**

```
if(!nearestVisibleAncestor) continue; // A root — no upward edge.
```

> Unexpected use of continue statement

**`app/(os-layout)/_components/row/TaskListDayView.tsx:105`**

```
continue;
```

> Unexpected use of continue statement

**`app/(os-layout)/_components/row/TaskListDayView.tsx:109`**

```
continue;
```

> Unexpected use of continue statement

**`app/(os-layout)/_components/row/useTaskRowActions.tsx:156`**

```
if(kind === input.task.kind) continue;
```

> Unexpected use of continue statement

**`app/(os-layout)/_components/select/TasksSelectionPageEffects.tsx:83`**

```
if(!selectedTaskIds.has(id)) continue;
```

> Unexpected use of continue statement

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

