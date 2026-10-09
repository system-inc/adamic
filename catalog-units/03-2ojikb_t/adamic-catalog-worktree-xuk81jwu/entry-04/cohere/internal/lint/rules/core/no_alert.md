# `no-alert`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **6** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the use of `alert`, `confirm`, and `prompt`

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

6 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:627`**

```
if(!confirm(`Delete ${selectedTaskId}? This also removes its subtree.`)) return;
```

> Unexpected confirm

**`app/(os-layout)/_components/detail/TaskDetailStackPane.tsx:92`**

```
if(!confirm(`Delete ${properties.taskId}? This also removes its subtree.`)) return;
```

> Unexpected confirm

**`app/(os-layout)/_components/row/TaskProjectHeader.tsx:121`**

```
if(!confirm(`Delete project "${properties.title || projectId}"? This also removes its subtree.`)) return;
```

> Unexpected confirm

**`app/(os-layout)/_components/select/TasksSelectionActionBar.tsx:130`**

```
if(!confirm(`Delete ${count} task${count === 1 ? '' : 's'}? This also removes their subtrees.`)) {
```

> Unexpected confirm

**`app/(os-layout)/_components/select/TasksSelectionActionBar.tsx:148`**

```
const cascade = confirm(
```

> Unexpected confirm

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

