# `@typescript-eslint/no-unnecessary-type-assertion`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **475** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow type assertions that do not change the type of an expression

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

475 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:65`**

```
const data = (await response.json()) as TaskNudgeResponseInterface;
```

> This assertion is unnecessary since it does not change the type of the expression

**`app/(os-layout)/_components/detail/TaskDetailActions.tsx:79`**

```
return (await response.json()) as { outcome: string; summary?: string };
```

> This assertion is unnecessary since it does not change the type of the expression

**`app/(os-layout)/_components/detail/TaskDetailCommandRulingCard.tsx:81`**

```
const data = (await response.json()) as CommandPreviewResponseInterface;
```

> This assertion is unnecessary since it does not change the type of the expression

**`app/(os-layout)/_components/kingdom/OsKingdomActivity.tsx:175`**

```
const page = (await response.json()) as ActivityPageInterface;
```

> This assertion is unnecessary since it does not change the type of the expression

**`app/(os-layout)/_components/kingdom/OsKingdomActivity.tsx:236`**

```
const page = (await response.json()) as ActivityPageInterface;
```

> This assertion is unnecessary since it does not change the type of the expression

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

