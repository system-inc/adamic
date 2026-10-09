# `@typescript-eslint/no-invalid-void-type`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **18** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow `void` type outside of generic or return types

## Why this recommendation

Every sampled site is `useWriteRequest<void, {...}>`, matching that hook's own declared signature `useWriteRequest<TData, TVariables = void>` in NetworkService.ts, so the rule fights the library's intended API shape.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong Yes** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

18 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/FinanceFloorView.tsx:70`**

```
const designateRequest = networkService.useWriteRequest<void, { accountId: string; designated: boolean }>({
```

> void is only valid as a return type or generic type argument

**`app/(os-layout)/finance/_components/FinanceReviewView.tsx:78`**

```
void,
```

> void is only valid as a return type or generic type argument

**`app/(os-layout)/finance/_components/FinanceReviewView.tsx:86`**

```
void,
```

> void is only valid as a return type or generic type argument

**`app/(os-layout)/finance/_components/transactions/FinanceTransactionDetail.tsx:109`**

```
const commentRequest = networkService.useWriteRequest<void, { transactionId: string; content: string }>({
```

> void is only valid as a return type or generic type argument

**`app/(os-layout)/finance/_components/transactions/FinanceTransactionDetail.tsx:113`**

```
const commentDeleteRequest = networkService.useWriteRequest<void, { transactionId: string; commentId: string }>({
```

> void is only valid as a return type or generic type argument

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

