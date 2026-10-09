# `consistent-return`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **129** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require `return` statements to either always or never specify values

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

129 in the tree. Showing the first few.

**`app/(os-layout)/_components/detail/TaskDetailCommandRulingCard.tsx:92`**

```
return function() {
```

> Function expected no return value

**`app/(os-layout)/_components/detail/TaskDetailFocusOverlay.tsx:94`**

```
return function() {
```

> Function expected no return value

**`app/(os-layout)/_components/detail/TaskDetailProjects.tsx:113`**

```
return registerProjectRegion({ addToProjectId: structuralProjectId, element });
```

> Function expected no return value

**`app/(os-layout)/_components/detail/TaskDetailSubtasks.tsx:82`**

```
return registerReparentRegion({ reparentToTaskId: openTaskId, element });
```

> Function expected no return value

**`app/(os-layout)/_components/drag/TasksDragGhostClone.tsx:25`**

```
return function() {
```

> Function expected no return value

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

