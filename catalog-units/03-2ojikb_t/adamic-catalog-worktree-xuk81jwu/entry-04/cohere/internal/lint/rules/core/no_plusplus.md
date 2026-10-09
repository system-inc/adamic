# `no-plusplus`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **592** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the unary operators `++` and `--`

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

592 in the tree. Showing the first few.

**`app/(os-layout)/_components/drag/TasksDragContext.tsx:490`**

```
for(let index = targetGapIndex; index < ordered.length; index++) {
```

> Unary operator '++' used

**`app/(os-layout)/os/providers/OsProvidersPalette.ts:50`**

```
for(let index = 0; index < profileId.length; index++) {
```

> Unary operator '++' used

**`app/(os-layout)/os/providers/OsProvidersSystemCard.tsx:56`**

```
if(member.wakefulness === 'Thinking') counts.Thinking++;
```

> Unary operator '++' used

**`app/(os-layout)/os/providers/OsProvidersSystemCard.tsx:57`**

```
else if(member.wakefulness === 'Awake') counts.Awake++;
```

> Unary operator '++' used

**`app/(os-layout)/os/providers/OsProvidersSystemCard.tsx:58`**

```
else counts.Asleep++;
```

> Unary operator '++' used

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

