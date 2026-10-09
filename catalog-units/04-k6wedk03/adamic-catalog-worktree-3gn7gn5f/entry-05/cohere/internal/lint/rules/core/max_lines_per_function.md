# `max-lines-per-function`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1980** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a maximum number of lines of code in a function

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

1980 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:41`**

```
export function TaskNudgeButton(properties: {
```

> Function 'TaskNudgeButton' has too many lines (73). Maximum allowed is 50

**`app/(os-layout)/_components/TasksCenter.tsx:149`**

```
export function TasksCenter() {
```

> Function 'TasksCenter' has too many lines (879). Maximum allowed is 50

**`app/(os-layout)/_components/TasksCenter.tsx:685`**

```
function(commit: TasksDragCommitInterface) {
```

> Function has too many lines (158). Maximum allowed is 50

**`app/(os-layout)/_components/TasksCommandMenu.tsx:32`**

```
export function TasksCommandMenu() {
```

> Function 'TasksCommandMenu' has too many lines (68). Maximum allowed is 50

**`app/(os-layout)/_components/TasksNewTaskDialog.tsx:68`**

```
export function TasksNewTaskDialog(properties: TasksNewTaskDialogProperties) {
```

> Function 'TasksNewTaskDialog' has too many lines (139). Maximum allowed is 50

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

