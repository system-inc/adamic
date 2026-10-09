# `require-unicode-regexp`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1471** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce the use of `u` or `v` flag on regular expressions

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

1471 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/activity/OsKingdomActivityToolResult.tsx:17`**

```
const preview = (properties.entry.text ?? '').replace(/\s+/g, ' ').trim().slice(0, 80);
```

> Use the 'u' flag

**`app/(os-layout)/_components/row/TaskAssigneePickerContent.tsx:31`**

```
const looksLikeEmailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
```

> Use the 'u' flag

**`app/(os-layout)/_layout/OsShell.tsx:103`**

```
.replace(/^\/tasks\/?/, '')
```

> Use the 'u' flag

**`app/(os-layout)/contacts/[id]/ContactHumanizeToken.tsx:15`**

```
.replace(/([a-z0-9])([A-Z])/g, '$1 $2')
```

> Use the 'u' flag

**`app/(os-layout)/contacts/[id]/ContactHumanizeToken.tsx:16`**

```
.replace(/[_-]+/g, ' ')
```

> Use the 'u' flag

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

