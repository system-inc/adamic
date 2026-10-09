# `prefer-arrow-callback`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **5619** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require using arrow functions for callbacks

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

5619 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:42`**

```
{navigationLinks.map(function(navigationLink, navigationLinkIndex) {
```

> Unexpected function expression

**`app/(os-layout)/_components/TaskNudgeButton.tsx:54`**

```
async function() {
```

> Unexpected function expression

**`app/(os-layout)/_components/TasksCenter.tsx:122`**

```
const beforeIndex = orderedSiblings.findIndex(function(task) {
```

> Unexpected function expression

**`app/(os-layout)/_components/TasksCenter.tsx:160`**

```
function() {
```

> Unexpected function expression

**`app/(os-layout)/_components/TasksCenter.tsx:222`**

```
function() {
```

> Unexpected function expression

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

