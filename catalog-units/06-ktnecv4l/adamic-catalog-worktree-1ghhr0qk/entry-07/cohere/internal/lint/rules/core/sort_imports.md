# `sort-imports`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **4726** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce sorted `import` declarations within modules

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

4726 in the tree. Showing the first few.

**`ProjectSettings.tsx:8`**

```
import AhraOsIcon from '@project/app/_assets/icons/AhraOsIcon.svg';
```

> Imports should be sorted alphabetically

**`app/(main-layout)/MainLayout.tsx:5`**

```
import { Navigation } from '@project/app/(main-layout)/_layout/navigation/Navigation';
```

> Imports should be sorted alphabetically

**`app/(main-layout)/MainLayout.tsx:8`**

```
import { LineLoadingAnimation } from '@structure/source/components/animations/LineLoadingAnimation';
```

> Imports should be sorted alphabetically

**`app/(main-layout)/_layout/navigation/Navigation.tsx:7`**

```
import { AppearanceSwitch } from '@structure/source/appearance/components/AppearanceSwitch';
```

> Imports should be sorted alphabetically

**`app/(main-layout)/_layout/navigation/Navigation.tsx:11`**

```
import { Link } from '@structure/source/components/navigation/Link';
```

> Imports should be sorted alphabetically

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

