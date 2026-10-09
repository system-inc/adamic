# `no-unneeded-ternary`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **9** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow ternary operators when simpler alternatives exist

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

9 in the tree. Showing the first few.

**`libraries/structure/source/components/dialogs/DialogRoot.tsx:175`**

```
(properties.footer === undefined &&
```

> Unnecessary use of boolean literals in conditional expression

**`libraries/structure/source/components/drawers/DrawerRoot.tsx:138`**

```
propertiesTransitionDurationInMilliseconds !== undefined || propertiesTransitionEasing !== undefined
```

> Unnecessary use of boolean literals in conditional expression

**`libraries/structure/source/components/forms/FormInstance.ts:289`**

```
isSubmitting: !hasErrors && live.onSubmitHandler ? true : false,
```

> Unnecessary use of boolean literals in conditional expression

**`libraries/structure/source/components/menus/Menu.tsx:137`**

```
const [loadingItems, setLoadingItems] = React.useState<boolean>(properties.isLoadingItems ? true : false);
```

> Unnecessary use of boolean literals in conditional expression

**`libraries/structure/source/components/tables/hooks/useTable.ts:513`**

```
next[columnId] = previous[columnId] === false ? true : false;
```

> Unnecessary use of boolean literals in conditional expression

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

