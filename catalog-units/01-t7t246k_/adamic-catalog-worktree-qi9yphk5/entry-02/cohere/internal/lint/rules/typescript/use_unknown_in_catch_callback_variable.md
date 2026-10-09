# `@typescript-eslint/use-unknown-in-catch-callback-variable`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **39** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Enforce typing arguments in Promise rejection callbacks as `unknown`

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

39 in the tree. Showing the first few.

**`app/(os-layout)/_components/row/useTaskRowActions.tsx:98`**

```
void taskDeleteRequest.execute({ id: input.task.id }).catch(function(error) {
```

> Prefer the safe `: unknown` for a `catch` callback variable

**`app/(os-layout)/art/_components/ArtCard.tsx:194`**

```
handleRatingChange(newRating).catch(function(error: Error) {
```

> Prefer the safe `: unknown` for a `catch` callback variable

**`app/(os-layout)/finance/_components/transactions/FinanceTransactionDetail.tsx:147`**

```
commentRequest.execute({ transactionId, content }).catch(function(error: Error) {
```

> Prefer the safe `: unknown` for a `catch` callback variable

**`app/(os-layout)/finance/_components/transactions/FinanceTransactionDetail.tsx:153`**

```
commentDeleteRequest.execute({ transactionId, commentId }).catch(function(error: Error) {
```

> Prefer the safe `: unknown` for a `catch` callback variable

**`app/(os-layout)/finance/_components/transactions/FinanceTransactionDetail.tsx:159`**

```
attachmentUploadRequest.execute({ transactionId, file }).catch(function(error: Error) {
```

> Prefer the safe `: unknown` for a `catch` callback variable

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

