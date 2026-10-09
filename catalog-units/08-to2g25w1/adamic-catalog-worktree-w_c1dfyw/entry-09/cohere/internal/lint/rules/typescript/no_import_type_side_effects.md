# `@typescript-eslint/no-import-type-side-effects`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **54** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce the use of top-level import type qualifier when an import only has specifiers with inline type qualifiers

## Why this recommendation

It directly reinforces the already-configured `consistent-type-imports` with `fixStyle: separate-type-imports`; 2304 imports already use `import type {` and only 59 inline stragglers remain, so it is an autofixable finisher for a convention already chosen.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Strong Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

54 in the tree. Showing the first few.

**`app/(os-layout)/_components/row/TaskRowMenu.tsx:20`**

```
import { type TaskInterface, type TaskPriorityKindType } from '@project/modules/tasks/Task';
```

> TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import

**`app/(os-layout)/_components/row/TaskRowMenu.tsx:22`**

```
import { type MenuItemProperties } from '@structure/source/components/menus/Menu';
```

> TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import

**`app/(os-layout)/_components/row/useTaskRowActions.tsx:29`**

```
import { type MenuItemProperties } from '@structure/source/components/menus/Menu';
```

> TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import

**`app/(os-layout)/_components/select/TasksSelectionActionBar.tsx:37`**

```
import { type TaskPriorityKindType, type TaskStatusKindType } from '@project/modules/tasks/Task';
```

> TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import

**`app/(os-layout)/_components/select/TasksSelectionActionBar.tsx:41`**

```
import { type MenuItemProperties } from '@structure/source/components/menus/Menu';
```

> TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

