# `react/jsx-sort-props`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **8336** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce props alphabetical sorting

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

8336 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:46`**

```
href={navigationLink.href}
```

> Props should be sorted alphabetically

**`app/(main-layout)/_layout/navigation/Navigation.tsx:47`**

```
className="p-2 text-sm font-medium transition-opacity hover:opacity-70"
```

> Props should be sorted alphabetically

**`app/(main-layout)/_layout/navigation/Navigation.tsx:58`**

```
<Link href="/" aria-label="Ahra Home" className="relative block size-8 shrink-0">
```

> Props should be sorted alphabetically

**`app/(main-layout)/_layout/navigation/Navigation.tsx:58`**

```
<Link href="/" aria-label="Ahra Home" className="relative block size-8 shrink-0">
```

> Props should be sorted alphabetically

**`app/(main-layout)/_layout/navigation/Navigation.tsx:64`**

```
<Link href="/" aria-label="Ahra Home" className="relative block size-8 shrink-0">
```

> Props should be sorted alphabetically

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

