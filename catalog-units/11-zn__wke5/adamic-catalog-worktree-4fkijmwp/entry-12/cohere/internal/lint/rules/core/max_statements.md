# `max-statements`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **2555** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a maximum number of statements allowed in function blocks

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

2555 in the tree. Showing the first few.

**`ProjectRoot.ts:127`**

```
function resolveProjectRoot(): string {
```

> Function 'resolveProjectRoot' has too many statements (11). Maximum allowed is 10

**`app/(os-layout)/_components/TaskNudgeButton.tsx:54`**

```
async function() {
```

> Async function has too many statements (14). Maximum allowed is 10

**`app/(os-layout)/_components/TasksCenter.tsx:102`**

```
function insertionGapBounds(
```

> Function 'insertionGapBounds' has too many statements (18). Maximum allowed is 10

**`app/(os-layout)/_components/TasksCenter.tsx:149`**

```
export function TasksCenter() {
```

> Function 'TasksCenter' has too many statements (87). Maximum allowed is 10

**`app/(os-layout)/_components/TasksCenter.tsx:392`**

```
const titleSegments: string[] = (function() {
```

> Function has too many statements (18). Maximum allowed is 10

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

