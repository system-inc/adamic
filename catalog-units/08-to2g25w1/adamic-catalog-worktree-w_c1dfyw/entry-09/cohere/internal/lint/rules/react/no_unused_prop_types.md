# `react/no-unused-prop-types`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **4** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow definitions of unused propTypes

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

4 in the tree. Showing the first few.

**`libraries/structure/source/components/images/editor/ImageEditor.tsx:55`**

```
minimumWidth?: number;
```

> 'minimumWidth' PropType is defined but prop is never used

**`libraries/structure/source/components/images/editor/ImageEditor.tsx:56`**

```
maximumWidth?: number;
```

> 'maximumWidth' PropType is defined but prop is never used

**`libraries/structure/source/components/images/editor/ImageEditor.tsx:57`**

```
minimumHeight?: number;
```

> 'minimumHeight' PropType is defined but prop is never used

**`libraries/structure/source/components/images/editor/ImageEditor.tsx:58`**

```
maximumHeight?: number;
```

> 'maximumHeight' PropType is defined but prop is never used

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

