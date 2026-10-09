# `react/jsx-boolean-value`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **101** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce boolean attributes notation in JSX

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

101 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:74`**

```
<AppearanceSwitch animated={true} />
```

> Value must be omitted for boolean attribute `animated`

**`app/(os-layout)/_components/TasksCenter.tsx:910`**

```
isOpen={true}
```

> Value must be omitted for boolean attribute `isOpen`

**`app/(os-layout)/_components/TasksNewTaskDialog.tsx:187`**

```
multiple={true}
```

> Value must be omitted for boolean attribute `multiple`

**`app/(os-layout)/_components/detail/TaskDetailBody.tsx:81`**

```
isDetailPanelOpen={true}
```

> Value must be omitted for boolean attribute `isDetailPanelOpen`

**`app/(os-layout)/_components/detail/TaskDetailDescriptionEditor.tsx:109`**

```
spellCheck={true}
```

> Value must be omitted for boolean attribute `spellCheck`

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

