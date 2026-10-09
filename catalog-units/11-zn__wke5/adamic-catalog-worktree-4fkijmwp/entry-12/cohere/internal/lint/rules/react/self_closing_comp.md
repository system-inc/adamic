# `react/self-closing-comp`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **8** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow extra closing tags for components without children

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

8 in the tree. Showing the first few.

**`libraries/structure/source/components/code/json/JsonNode.tsx:247`**

```
<span
```

> Empty components are self-closing

**`libraries/structure/source/components/forms/fields/choice-grid/FieldInputChoiceGrid.tsx:46`**

```
<th className="p-2"></th>
```

> Empty components are self-closing

**`libraries/structure/source/components/forms/fields/multiple-checkbox-grid/FieldInputMultipleCheckboxGrid.tsx:69`**

```
<th className="p-2"></th>
```

> Empty components are self-closing

**`libraries/structure/source/layouts/side-navigation/SideNavigationLayoutNavigationSide.tsx:714`**

```
<div
```

> Empty components are self-closing

**`libraries/structure/source/modules/account/authentication/components/AuthenticationEmailForm.tsx:169`**

```
<div className="size-5 animate-spin rounded-full border-2 border--0 border-t-transparent"></div>
```

> Empty components are self-closing

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

