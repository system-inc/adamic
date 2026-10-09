# `no-void`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **461** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow `void` operators

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

461 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:101`**

```
void nudge();
```

> Expected 'undefined' and instead saw 'void'

**`app/(os-layout)/_components/TasksCenter.tsx:470`**

```
void setFocusParameter(null);
```

> Expected 'undefined' and instead saw 'void'

**`app/(os-layout)/_components/TasksCenter.tsx:471`**

```
void setStackParameter(null);
```

> Expected 'undefined' and instead saw 'void'

**`app/(os-layout)/_components/TasksCenter.tsx:474`**

```
void setFocusParameter('true');
```

> Expected 'undefined' and instead saw 'void'

**`app/(os-layout)/_components/TasksCenter.tsx:483`**

```
void setStackParameter(length > 0 ? stackTaskIds.slice(0, length).join(',') : null);
```

> Expected 'undefined' and instead saw 'void'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

