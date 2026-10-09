# `@typescript-eslint/prefer-readonly-parameter-types`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **12744** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Require function parameters to be typed as `readonly` to prevent accidental mutation of inputs

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

12744 in the tree. Showing the first few.

**`ProjectRoot.ts:173`**

```
export function projectPath(...pathSegments: string[]): string {
```

> Parameter should be a read only type

**`app/(main-layout)/MainLayout.tsx:14`**

```
export function MainLayout(properties: MainLayoutProperties) {
```

> Parameter should be a read only type

**`app/(main-layout)/_layout/navigation/Navigation.tsx:30`**

```
export function Navigation(properties: NavigationProperties) {
```

> Parameter should be a read only type

**`app/(main-layout)/_layout/navigation/Navigation.tsx:42`**

```
{navigationLinks.map(function(navigationLink, navigationLinkIndex) {
```

> Parameter should be a read only type

**`app/(os-layout)/AhraChatView.tsx:40`**

```
kingdomEventSourceSnapshot.data?.members.find((member) => member.username === DyadUsername) ?? null;
```

> Parameter should be a read only type

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

