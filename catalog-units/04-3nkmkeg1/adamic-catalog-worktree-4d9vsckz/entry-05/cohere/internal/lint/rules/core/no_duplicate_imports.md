# `no-duplicate-imports`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **610** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow duplicate module imports

## Why this recommendation

Nearly every hit is the deliberate `import type {...}` / `import {...}` split that the in-house consistent-type-imports rule produces, as in TasksCenter.tsx lines 42-43, so it would fight an existing rule rather than find defects.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

610 in the tree. Showing the first few.

**`app/(os-layout)/AhraChatView.tsx:20`**

```
import { DyadUsername } from '@project/modules/os/AhraOsKingdom';
```

> '@project/modules/os/AhraOsKingdom' import is duplicated

**`app/(os-layout)/_components/TasksCenter.tsx:43`**

```
import { TasksDragProvider } from './drag/TasksDragContext';
```

> './drag/TasksDragContext' import is duplicated

**`app/(os-layout)/_components/TasksCenter.tsx:70`**

```
import { TaskNodeKind } from '@project/modules/tasks/Task';
```

> '@project/modules/tasks/Task' import is duplicated

**`app/(os-layout)/_components/detail/TaskDetail.tsx:37`**

```
import { TaskDetailCommentComposer } from './TaskDetailCommentComposer';
```

> './TaskDetailCommentComposer' import is duplicated

**`app/(os-layout)/_components/detail/TaskDetail.tsx:43`**

```
import { motion } from 'motion/react';
```

> 'motion/react' import is duplicated

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

