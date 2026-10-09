# `@typescript-eslint/explicit-function-return-type`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **3786** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require explicit return types on functions and class methods

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

3786 in the tree. Showing the first few.

**`app/(main-layout)/MainLayout.tsx:14`**

```
export function MainLayout(properties: MainLayoutProperties) {
```

> Missing return type on function

**`app/(main-layout)/_layout/navigation/Navigation.tsx:30`**

```
export function Navigation(properties: NavigationProperties) {
```

> Missing return type on function

**`app/(os-layout)/AhraChatView.tsx:29`**

```
export function AhraChatView() {
```

> Missing return type on function

**`app/(os-layout)/_components/TaskNudgeButton.tsx:41`**

```
export function TaskNudgeButton(properties: {
```

> Missing return type on function

**`app/(os-layout)/_components/TasksCenter.tsx:149`**

```
export function TasksCenter() {
```

> Missing return type on function

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

