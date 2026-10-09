# `no-negated-condition`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **504** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow negated conditions

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

504 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:107`**

```
{nudgeNotice !== null ? (
```

> Unexpected negated condition

**`app/(os-layout)/_components/TaskNudgeButton.tsx:110`**

```
{nudgeError !== null ? <span className="text-xs text-red-600 dark:text-red-400">{nudgeError}</span> : null}
```

> Unexpected negated condition

**`app/(os-layout)/_components/TasksCenter.tsx:292`**

```
const scopedToCurrent = scopeParameter !== null ? scopeParameter === 'current' : projectId !== null;
```

> Unexpected negated condition

**`app/(os-layout)/_components/TasksCenter.tsx:331`**

```
const activeProject = projectId !== null ? findProject(projectId) : undefined;
```

> Unexpected negated condition

**`app/(os-layout)/_components/TasksCenter.tsx:332`**

```
const activeProjectTitle = projectId !== null ? activeProject?.title || projectId : undefined;
```

> Unexpected negated condition

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

