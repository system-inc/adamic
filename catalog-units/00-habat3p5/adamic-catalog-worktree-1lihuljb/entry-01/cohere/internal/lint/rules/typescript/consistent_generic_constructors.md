# `@typescript-eslint/consistent-generic-constructors`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **11** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce specifying generic type arguments on type annotation or constructor name of a constructor call

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

11 in the tree. Showing the first few.

**`app/_assets/icons/poring/PoringIcon.tsx:50`**

```
const weights: Map<IconWeight, React.ReactElement<unknown, string | React.JSXElementConstructor<unknown>>> = new Map([
```

> The generic type arguments should be specified as part of the constructor type arguments

**`libraries/structure/assets/icons/commerce/AmexIcon.tsx:9`**

```
const weights: Map<
```

> The generic type arguments should be specified as part of the constructor type arguments

**`libraries/structure/assets/icons/commerce/DiscoverIcon.tsx:8`**

```
const weights: Map<
```

> The generic type arguments should be specified as part of the constructor type arguments

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:8`**

```
const weights: Map<
```

> The generic type arguments should be specified as part of the constructor type arguments

**`libraries/structure/assets/icons/commerce/VisaIcon.tsx:8`**

```
const weights: Map<
```

> The generic type arguments should be specified as part of the constructor type arguments

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

