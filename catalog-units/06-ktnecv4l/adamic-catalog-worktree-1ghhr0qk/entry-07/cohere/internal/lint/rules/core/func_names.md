# `func-names`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **7717** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require or disallow named `function` expressions

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

7717 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:42`**

```
{navigationLinks.map(function(navigationLink, navigationLinkIndex) {
```

> Unexpected unnamed function

**`app/(os-layout)/_components/TaskNudgeButton.tsx:54`**

```
async function() {
```

> Unexpected unnamed async function

**`app/(os-layout)/_components/TaskNudgeButton.tsx:100`**

```
onClick={function() {
```

> Unexpected unnamed function

**`app/(os-layout)/_components/TasksCenter.tsx:122`**

```
const beforeIndex = orderedSiblings.findIndex(function(task) {
```

> Unexpected unnamed function

**`app/(os-layout)/_components/TasksCenter.tsx:141`**

```
positionFor: function(index: number, count: number): number {
```

> Unexpected unnamed method 'positionFor'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

