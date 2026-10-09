# `no-console`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **9565** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the use of `console`

## Why this recommendation

Formatting, which Prettier already decides. Turning it on creates a second authority that can disagree.

## Violations

9565 in the tree. Showing the first few.

**`app/(os-layout)/_components/TaskNudgeButton.tsx:78`**

```
console.error('Failed to nudge:', error);
```

> Unexpected console statement

**`app/(os-layout)/_components/detail/TaskDetailActions.tsx:82`**

```
console.error(`Failed to ${decision} verdict ${verdictId}:`, error);
```

> Unexpected console statement

**`app/(os-layout)/_components/detail/TaskDetailCommandRulingCard.tsx:88`**

```
console.error(`Failed to preview edited command for ${verdictId}:`, error);
```

> Unexpected console statement

**`app/(os-layout)/art/_components/ArtCard.tsx:195`**

```
console.error('Failed to save rating:', error);
```

> Unexpected console statement

**`app/(os-layout)/art/_components/UniverseWeightsDialogBody.tsx:101`**

```
console.error('Failed to save universe weights:', saveException);
```

> Unexpected console statement

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

