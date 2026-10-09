# `react/prefer-read-only-props`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **364** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce that props are read-only

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

364 in the tree. Showing the first few.

**`app/_assets/icons/AhraIcon.tsx:10`**

```
}: React.SVGProps<SVGSVGElement> & { ref?: React.Ref<SVGSVGElement> }) {
```

> Prop 'ref' should be read-only

**`libraries/structure/assets/icons/IconBase.tsx:11`**

```
weights: Map<IconWeightType, React.ReactElement>;
```

> Prop 'weights' should be read-only

**`libraries/structure/assets/icons/IconBase.tsx:12`**

```
ref?: React.Ref<SVGSVGElement>;
```

> Prop 'ref' should be read-only

**`libraries/structure/assets/icons/commerce/AmexIcon.tsx:31`**

```
}: IconProperties & { ref?: React.Ref<SVGSVGElement> }) {
```

> Prop 'ref' should be read-only

**`libraries/structure/assets/icons/commerce/DiscoverIcon.tsx:36`**

```
}: IconProperties & { ref?: React.Ref<SVGSVGElement> }) {
```

> Prop 'ref' should be read-only

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

