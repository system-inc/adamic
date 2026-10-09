# `no-nested-ternary`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **459** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow nested ternary expressions

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

459 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:335`**

```
const listTitle = isSearching
```

> Do not nest ternary expressions

**`app/(os-layout)/_components/TasksCenter.tsx:344`**

```
const listSubtitle = isSearching
```

> Do not nest ternary expressions

**`app/(os-layout)/_components/TasksCenter.tsx:350`**

```
const listIsLoading = isSearching
```

> Do not nest ternary expressions

**`app/(os-layout)/_components/detail/TaskDetail.tsx:146`**

```
{properties.task ? (
```

> Do not nest ternary expressions

**`app/(os-layout)/_components/detail/TaskDetailCommandFiredCard.tsx:67`**

```
{succeeded ? 'ok' : exitCode !== null ? `exit ${exitCode}` : 'failed'}
```

> Do not nest ternary expressions

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

