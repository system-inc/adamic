# `@typescript-eslint/no-unnecessary-condition`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **1243** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow conditionals where the type is always truthy or always falsy

## Why this recommendation

Catches a defect rather than a preference, but the volume means it needs a plan rather than a switch.

## Violations

1243 in the tree. Showing the first few.

**`ProjectRoot.ts:86`**

```
while(true) {
```

> Unnecessary conditional, value is always truthy

**`app/(os-layout)/_components/TasksCenter.tsx:158`**

```
const slug = urlParameters?.slug as string[] | undefined;
```

> Unnecessary optional chain on a non-nullish value

**`app/(os-layout)/_components/TasksCenter.tsx:170`**

```
const currentQueryString = urlSearchParameters?.toString() ?? '';
```

> Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined

**`app/(os-layout)/_components/TasksCenter.tsx:170`**

```
const currentQueryString = urlSearchParameters?.toString() ?? '';
```

> Unnecessary optional chain on a non-nullish value

**`app/(os-layout)/_components/TasksCenter.tsx:256`**

```
const projectId = typeof selection === 'object' && selection.kind === 'Project' ? selection.id : null;
```

> Unnecessary conditional, comparison is always true, since `"Project" === "Project"` is true

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

