# `@typescript-eslint/array-type`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **711** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require consistently using either `T[]` or `Array<T>` for arrays

## Why this recommendation

Not a house-style conflict: the codebase already writes `readonly T[]` 213 times against only 46 `ReadonlyArray<>`, so the rule's default matches existing practice and is fully autofixable.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

711 in the tree. Showing the first few.

**`app/(os-layout)/_components/select/TasksSelectionContext.tsx:30`**

```
replace: (taskIds: ReadonlyArray<string>, anchorTaskId?: string | null) => void;
```

> Array type using 'ReadonlyArray<string>' is forbidden. Use 'readonly string[]' instead

**`app/(os-layout)/_components/select/TasksSelectionContext.tsx:31`**

```
add: (taskIds: ReadonlyArray<string>) => void;
```

> Array type using 'ReadonlyArray<string>' is forbidden. Use 'readonly string[]' instead

**`app/(os-layout)/_components/select/TasksSelectionContext.tsx:32`**

```
remove: (taskIds: ReadonlyArray<string>) => void;
```

> Array type using 'ReadonlyArray<string>' is forbidden. Use 'readonly string[]' instead

**`app/(os-layout)/_components/select/TasksSelectionContext.tsx:92`**

```
const replace = React.useCallback(function(taskIds: ReadonlyArray<string>, nextAnchorTaskId: string | null = null) {
```

> Array type using 'ReadonlyArray<string>' is forbidden. Use 'readonly string[]' instead

**`app/(os-layout)/_components/select/TasksSelectionContext.tsx:98`**

```
const add = React.useCallback(function(taskIds: ReadonlyArray<string>) {
```

> Array type using 'ReadonlyArray<string>' is forbidden. Use 'readonly string[]' instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

