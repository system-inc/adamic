# `react/jsx-no-literals`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1785** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow usage of string literals in JSX

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

1785 in the tree. Showing the first few.

**`app/(os-layout)/AhraChatView.tsx:49`**

```
<span className="text-sm font-semibold">Ahra</span>
```

> Missing JSX expression container around literal string: "Ahra"

**`app/(os-layout)/AhraChatView.tsx:50`**

```
<span className="content--4">·</span>
```

> Missing JSX expression container around literal string: "·"

**`app/(os-layout)/AhraChatView.tsx:51`**

```
<span className="text-sm content--3">Dyad</span>
```

> Missing JSX expression container around literal string: "Dyad"

**`app/(os-layout)/_components/TaskNudgeButton.tsx:104`**

```
>
```

> Missing JSX expression container around literal string: "Nudge"

**`app/(os-layout)/_components/TasksNewTaskDialog.tsx:190`**

```
<div className="text-[11px] content--4">drops into {destinationLabel}</div>
```

> Missing JSX expression container around literal string: "drops into"

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

