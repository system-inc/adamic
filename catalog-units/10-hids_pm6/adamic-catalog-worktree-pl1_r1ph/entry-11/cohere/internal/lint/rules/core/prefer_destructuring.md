# `prefer-destructuring`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **789** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require destructuring from arrays and/or objects

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

789 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:171`**

```
const selection = urlState.selection;
```

> Use object destructuring

**`app/(os-layout)/_components/TasksCenter.tsx:172`**

```
const selectedTaskId = urlState.selectedTaskId;
```

> Use object destructuring

**`app/(os-layout)/_components/detail/TaskDetailDescriptionEditor.tsx:36`**

```
const taskId = properties.taskId;
```

> Use object destructuring

**`app/(os-layout)/_components/detail/TaskDetailFocusOverlay.tsx:70`**

```
const onExit = properties.onExit;
```

> Use object destructuring

**`app/(os-layout)/_components/detail/TaskDetailProjects.tsx:105`**

```
const registerProjectRegion = tasksDrag.registerProjectRegion;
```

> Use object destructuring

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

