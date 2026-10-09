# `react/jsx-no-useless-fragment`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **18** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow unnecessary fragments

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

18 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/statements/FinanceStatementPayeeGroupsBody.tsx:35`**

```
<>
```

> Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all

**`app/(os-layout)/finance/_components/statements/FinanceStatementTransactionsBody.tsx:29`**

```
<>
```

> Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all

**`app/(os-layout)/os/wisdom/_components/WisdomStoryText.tsx:105`**

```
return <>{properties.children}</>;
```

> Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all

**`libraries/structure/assets/icons/commerce/AmexIcon.tsx:15`**

```
<>
```

> Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all

**`libraries/structure/assets/icons/commerce/VisaIcon.tsx:14`**

```
<>
```

> Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

