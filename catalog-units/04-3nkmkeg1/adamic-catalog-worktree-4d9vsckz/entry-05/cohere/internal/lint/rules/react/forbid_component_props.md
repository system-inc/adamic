# `react/forbid-component-props`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1357** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow certain props on components

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

1357 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:47`**

```
className="p-2 text-sm font-medium transition-opacity hover:opacity-70"
```

> Prop "className" is forbidden on Components

**`app/(main-layout)/_layout/navigation/Navigation.tsx:58`**

```
<Link href="/" aria-label="Ahra Home" className="relative block size-8 shrink-0">
```

> Prop "className" is forbidden on Components

**`app/(main-layout)/_layout/navigation/Navigation.tsx:59`**

```
<ProjectIcon className="size-8" />
```

> Prop "className" is forbidden on Components

**`app/(main-layout)/_layout/navigation/Navigation.tsx:64`**

```
<Link href="/" aria-label="Ahra Home" className="relative block size-8 shrink-0">
```

> Prop "className" is forbidden on Components

**`app/(main-layout)/_layout/navigation/Navigation.tsx:65`**

```
<ProjectIcon className="size-8" />
```

> Prop "className" is forbidden on Components

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

