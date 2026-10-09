# `no-ternary`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **7737** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow ternary operators

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

7737 in the tree. Showing the first few.

**`ProjectRoot.ts:72`**

```
const thisFileDirectory: string | undefined = import.meta.url
```

> Ternary operator used

**`app/(os-layout)/_components/TaskNudgeButton.tsx:72`**

```
data.ownerNotified && data.ownerHandle
```

> Ternary operator used

**`app/(os-layout)/_components/TaskNudgeButton.tsx:107`**

```
{nudgeNotice !== null ? (
```

> Ternary operator used

**`app/(os-layout)/_components/TaskNudgeButton.tsx:110`**

```
{nudgeError !== null ? <span className="text-xs text-red-600 dark:text-red-400">{nudgeError}</span> : null}
```

> Ternary operator used

**`app/(os-layout)/_components/TasksCenter.tsx:127`**

```
const predecessor = beforeIndex > 0 ? orderedSiblings[beforeIndex - 1] : undefined;
```

> Ternary operator used

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

