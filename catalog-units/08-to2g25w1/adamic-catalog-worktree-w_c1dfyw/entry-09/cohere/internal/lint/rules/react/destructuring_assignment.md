# `react/destructuring-assignment`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **6869** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce consistent usage of destructuring assignment of props, state, and context

## Why this recommendation

It does not merely overlap the in-house rules, it directly contradicts them: react-component-no-destructuring and react-hook-no-destructuring ban destructuring, and every one of the 6869 hits is the house `properties.children` / `properties.className` style the CLAUDE.md 'Reach Over Alias' convention mandates.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Strong No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

6869 in the tree. Showing the first few.

**`app/(main-layout)/MainLayout.tsx:32`**

```
<div className="relative w-full overflow-x-clip">{properties.children}</div>
```

> Must use destructuring properties assignment

**`app/(main-layout)/_layout/navigation/Navigation.tsx:36`**

```
properties.className,
```

> Must use destructuring properties assignment

**`app/(os-layout)/_components/TaskNudgeButton.tsx:51`**

```
const propertiesTaskId = properties.taskId;
```

> Must use destructuring properties assignment

**`app/(os-layout)/_components/TaskNudgeButton.tsx:93`**

```
<span className={`inline-flex items-center gap-2 ${properties.className ?? ''}`}>
```

> Must use destructuring properties assignment

**`app/(os-layout)/_components/TasksNewTaskDialog.tsx:81`**

```
const propertiesOnClose = properties.onClose;
```

> Must use destructuring properties assignment

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

