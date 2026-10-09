# `no-inline-comments`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **3976** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow inline comments after code

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

3976 in the tree. Showing the first few.

**`app/(main-layout)/_layout/navigation/Navigation.tsx:1`**

```
'use client'; // Uses client-only features
```

> Unexpected comment inline with code

**`app/(os-layout)/AhraChatView.tsx:1`**

```
'use client'; // Uses client-only features
```

> Unexpected comment inline with code

**`app/(os-layout)/_components/TaskNudgeButton.tsx:1`**

```
'use client'; // Uses client-only features
```

> Unexpected comment inline with code

**`app/(os-layout)/_components/TasksCenter.tsx:1`**

```
'use client'; // Uses client-only features
```

> Unexpected comment inline with code

**`app/(os-layout)/_components/TasksCommandMenu.tsx:1`**

```
'use client'; // Uses client-only features
```

> Unexpected comment inline with code

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

