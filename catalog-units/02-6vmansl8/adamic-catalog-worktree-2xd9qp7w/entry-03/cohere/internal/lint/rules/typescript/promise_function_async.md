# `@typescript-eslint/promise-function-async`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **367** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Require any function or method that returns a Promise to be marked async

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

367 in the tree. Showing the first few.

**`app/(os-layout)/_hooks/useMessagesFeedRequest.ts:201`**

```
.then(function(response) {
```

> Functions that return promises must be async

**`app/(os-layout)/finance/_components/FinanceContactRow.tsx:223`**

```
<>{properties.rows.map((row) => renderTransactionRow(row))}</>
```

> Functions that return promises must be async. Consider adding an explicit return type annotation if the function is intended to return a union of promise and non-promise types

**`app/(os-layout)/finance/_components/transactions/FinanceTransactionYearGroups.tsx:124`**

```
{properties.group.rows.map((row) => properties.renderRow(row))}
```

> Functions that return promises must be async. Consider adding an explicit return type annotation if the function is intended to return a union of promise and non-promise types

**`app/(os-layout)/os/wisdom/_components/WisdomBundleChecklist.tsx:82`**

```
function(stepId: string, action: 'approve' | 'reject', edit?: { editedText?: string; editedCommand?: string }) {
```

> Functions that return promises must be async

**`app/(os-layout)/os/wisdom/_components/WisdomBundleChecklist.tsx:83`**

```
return wisdomPending.run(function() {
```

> Functions that return promises must be async

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

