# `prefer-template`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **707** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require template literals instead of string concatenation

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

707 in the tree. Showing the first few.

**`app/(os-layout)/_components/detail/TaskDetailCommandFiredCard.tsx:61`**

```
'shrink-0 rounded-sm px-1.5 py-0.5 text-[9.5px] font-semibold uppercase ' +
```

> Unexpected string concatenation

**`app/(os-layout)/_components/detail/TaskDetailCommandOutput.tsx:22`**

```
'shrink-0 rounded-sm px-1.5 py-0.5 text-[9.5px] font-semibold uppercase ' +
```

> Unexpected string concatenation

**`app/(os-layout)/_layout/sidebar/Sidebar.tsx:269`**

```
<div className="flex size-full shrink-0 flex-col" style={{ minWidth: innerMinimumWidth + 'px' }}>
```

> Unexpected string concatenation

**`app/(os-layout)/_layout/sidebar/SidebarSectionHeader.tsx:13`**

```
'overflow-hidden px-2 pt-3 pb-1 text-xs whitespace-nowrap content--3 uppercase transition-all ' +
```

> Unexpected string concatenation

**`app/(os-layout)/finance/_components/FinanceSidebar.tsx:129`**

```
const matches = urlPath === href || urlPath.startsWith(href + '/');
```

> Unexpected string concatenation

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

