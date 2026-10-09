# `react/jsx-handler-names`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **299** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce event handler naming conventions in JSX

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

299 in the tree. Showing the first few.

**`app/(os-layout)/_components/detail/TaskDetail.tsx:134`**

```
onFocusViewModeChange={taskDetailFocusViewMode.setFocusViewMode}
```

> Handler function for onFocusViewModeChange prop key must be a camelCase name beginning with 'handle' only

**`app/(os-layout)/_components/detail/TaskDetail.tsx:135`**

```
onStatusChange={properties.onStatusChange}
```

> Handler function for onStatusChange prop key must be a camelCase name beginning with 'handle' only

**`app/(os-layout)/_components/detail/TaskDetail.tsx:136`**

```
onToggleFocusMode={properties.onToggleFocusMode}
```

> Handler function for onToggleFocusMode prop key must be a camelCase name beginning with 'handle' only

**`app/(os-layout)/_components/detail/TaskDetail.tsx:137`**

```
onClose={properties.onClose}
```

> Handler function for onClose prop key must be a camelCase name beginning with 'handle' only

**`app/(os-layout)/_components/detail/TaskDetail.tsx:158`**

```
onOpenSubtask={properties.onOpenSubtask}
```

> Handler function for onOpenSubtask prop key must be a camelCase name beginning with 'handle' only

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

