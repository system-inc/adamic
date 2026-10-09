# `react/jsx-curly-brace-presence`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **11** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow unnecessary JSX expressions when literals alone are sufficient or enforce JSX expressions on literals in JSX children or attributes

## Why this recommendation

Prettier does not add or strip JSX curly braces around string literals, so there is no competing authority, and at only 11 auto-fixable sites it is a cheap consistency win.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

11 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/statements/FinanceStatementPrimitives.tsx:139`**

```
<span>In {'·'} Out</span>
```

> Curly braces are unnecessary here

**`app/(os-layout)/os/sensation/_components/SensationAwaitingReactionRow.tsx:91`**

```
className={
```

> Curly braces are unnecessary here

**`libraries/structure/source/components/controls/switch/Switch.tsx:113`**

```
whileTap={'pressed'}
```

> Curly braces are unnecessary here

**`libraries/structure/source/components/interactions/Collapse.tsx:31`**

```
className={'relative w-full overflow-x-auto overflow-y-hidden'}
```

> Curly braces are unnecessary here

**`libraries/structure/source/components/notifications/NotificationsContainer.tsx:211`**

```
exit={'exit'}
```

> Curly braces are unnecessary here

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

