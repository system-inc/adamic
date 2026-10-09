# `react/prop-types`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **5** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow missing props validation in a React component definition

## Why this recommendation

It is not an overlap but a category error in a TypeScript codebase: the 5 hits ask for runtime propTypes on components whose props are already typed, such as AhraIcon.tsx's `width`/`height`.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Strong No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

5 in the tree. Showing the first few.

**`app/_assets/icons/AhraIcon.tsx:7`**

```
width = 24,
```

> 'width' is missing in props validation

**`app/_assets/icons/AhraIcon.tsx:8`**

```
height = 24,
```

> 'height' is missing in props validation

**`libraries/structure/source/components/tables/parts/TableBody.tsx:32`**

```
export function TableBody({ className, style, children, ...tbodyProperties }: TableBodyProperties) {
```

> 'style' is missing in props validation

**`libraries/structure/source/components/tables/parts/TableCell.tsx:19`**

```
export function TableCell({ className, ...tdProperties }: TableCellProperties) {
```

> 'className' is missing in props validation

**`libraries/structure/source/components/tables/parts/TableHeader.tsx:29`**

```
export function TableHeader({ className, style, ...theadProperties }: TableHeaderProperties) {
```

> 'style' is missing in props validation

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

