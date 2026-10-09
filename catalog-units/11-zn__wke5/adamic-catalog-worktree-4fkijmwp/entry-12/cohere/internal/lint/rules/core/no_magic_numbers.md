# `no-magic-numbers`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **22253** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow magic numbers

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

22253 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:55`**

```
if(isNudging || !propertiesTaskId || propertiesTaskId.length === 0) {
```

> No magic number: 0

**`app/(os-layout)/_components/TaskNudgeButton.tsx:88`**

```
if(!propertiesTaskId || propertiesTaskId.length === 0) {
```

> No magic number: 0

**`app/(os-layout)/_components/TasksCenter.tsx:111`**

```
const lastSibling = orderedSiblings.at(-1);
```

> No magic number: -1

**`app/(os-layout)/_components/TasksCenter.tsx:114`**

```
lowerBound = 0;
```

> No magic number: 0

**`app/(os-layout)/_components/TasksCenter.tsx:127`**

```
const predecessor = beforeIndex > 0 ? orderedSiblings[beforeIndex - 1] : undefined;
```

> No magic number: 0

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

