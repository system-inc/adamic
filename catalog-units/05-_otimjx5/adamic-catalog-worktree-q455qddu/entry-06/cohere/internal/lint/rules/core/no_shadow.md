# `no-shadow`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **174** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow variable declarations from shadowing variables declared in the outer scope

## Why this recommendation

@typescript-eslint/no-shadow is already recommended at the identical count of 174, and the base rule mis-handles TypeScript type and enum declaration merging, so enabling both double-reports every site.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Yes** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

174 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:437`**

```
function handleSelectTask(taskId: string) {
```

> 'handleSelectTask' is already declared in the upper scope on line 436 column 11

**`app/(os-layout)/_components/TasksCenter.tsx:458`**

```
function handleCloseDrawer() {
```

> 'handleCloseDrawer' is already declared in the upper scope on line 457 column 11

**`app/(os-layout)/_components/TasksCenter.tsx:465`**

```
function handleToggleFocusMode() {
```

> 'handleToggleFocusMode' is already declared in the upper scope on line 464 column 11

**`app/(os-layout)/_components/TasksCenter.tsx:482`**

```
function handlePopStackToLength(length: number) {
```

> 'handlePopStackToLength' is already declared in the upper scope on line 481 column 11

**`app/(os-layout)/_components/TasksCenter.tsx:492`**

```
function handlePushStackTask(taskId: string) {
```

> 'handlePushStackTask' is already declared in the upper scope on line 491 column 11

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

