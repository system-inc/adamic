# `react/require-default-props`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **337** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a defaultProps definition for every prop that is not a required prop

## Why this recommendation

defaultProps is deprecated and removed for function components in React 19, so this rule pushes the codebase toward a dead API rather than merely duplicating in-house coverage.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Strong No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

337 in the tree. Showing the first few.

**`app/_assets/icons/AhraIcon.tsx:10`**

```
}: React.SVGProps<SVGSVGElement> & { ref?: React.Ref<SVGSVGElement> }) {
```

> propType "ref" is not required, but has no corresponding defaultProps declaration

**`libraries/structure/assets/icons/IconBase.tsx:12`**

```
ref?: React.Ref<SVGSVGElement>;
```

> propType "ref" is not required, but has no corresponding defaultProps declaration

**`libraries/structure/assets/icons/commerce/AmexIcon.tsx:31`**

```
}: IconProperties & { ref?: React.Ref<SVGSVGElement> }) {
```

> propType "ref" is not required, but has no corresponding defaultProps declaration

**`libraries/structure/assets/icons/commerce/DiscoverIcon.tsx:36`**

```
}: IconProperties & { ref?: React.Ref<SVGSVGElement> }) {
```

> propType "ref" is not required, but has no corresponding defaultProps declaration

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:41`**

```
}: IconProperties & { ref?: React.Ref<SVGSVGElement> }) {
```

> propType "ref" is not required, but has no corresponding defaultProps declaration

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

