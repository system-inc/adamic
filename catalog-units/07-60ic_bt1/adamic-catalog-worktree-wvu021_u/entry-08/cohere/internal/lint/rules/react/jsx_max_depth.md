# `react/jsx-max-depth`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1111** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce JSX maximum depth

## Why this recommendation

It is an arbitrary structural-complexity cap, not formatting, so Prettier is not the reason to decline; the real reason is that 1111 hits mean the codebase has settled on deeper JSX trees.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

1111 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:59`**

```
<ProjectIcon className="size-8" />
```

> Expected the depth of nested jsx elements to be <= 2, but found 3

**`app/(os-layout)/_components/TasksCenter.tsx:899`**

```
<TasksDragCommitBridge onCommit={handleDragCommit} />
```

> Expected the depth of nested jsx elements to be <= 2, but found 3

**`app/(os-layout)/_components/TasksCenter.tsx:900`**

```
<TasksDragGhost />
```

> Expected the depth of nested jsx elements to be <= 2, but found 3

**`app/(os-layout)/_components/TasksCenter.tsx:901`**

```
<TasksDragInsertionLine />
```

> Expected the depth of nested jsx elements to be <= 2, but found 3

**`app/(os-layout)/_components/TasksCenter.tsx:902`**

```
<TasksSelectionPageEffects listedTasks={listedTasks} />
```

> Expected the depth of nested jsx elements to be <= 2, but found 3

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

