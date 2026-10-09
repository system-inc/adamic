# `react/jsx-no-bind`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1063** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow `.bind()` or arrow functions in JSX props

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

1063 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:100`**

```
onClick={function() {
```

> JSX props should not use functions

**`app/(os-layout)/_components/TasksCenter.tsx:866`**

```
onPriorityChange={function(priority) {
```

> JSX props should not use functions

**`app/(os-layout)/_components/TasksCenter.tsx:869`**

```
onStatusChange={function(status) {
```

> JSX props should not use functions

**`app/(os-layout)/_components/TasksCenter.tsx:872`**

```
onAddComment={function(content) {
```

> JSX props should not use functions

**`app/(os-layout)/_components/TasksCenter.tsx:880`**

```
onDismissCommentWakeOutcome={function() {
```

> JSX props should not use functions

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

