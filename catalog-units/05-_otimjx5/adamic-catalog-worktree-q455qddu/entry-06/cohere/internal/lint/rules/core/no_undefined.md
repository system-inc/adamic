# `no-undefined`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **3235** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the use of `undefined` as an identifier

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

3235 in the tree. Showing the first few.

**`ProjectRoot.ts:75`**

```
if(!thisFileDirectory) return undefined;
```

> Unexpected use of undefined

**`ProjectRoot.ts:89`**

```
if(parentDirectory === currentDirectory) return undefined;
```

> Unexpected use of undefined

**`app/(os-layout)/_components/TasksCenter.tsx:113`**

```
if(lastSibling === undefined || headSibling === undefined) {
```

> Unexpected use of undefined

**`app/(os-layout)/_components/TasksCenter.tsx:113`**

```
if(lastSibling === undefined || headSibling === undefined) {
```

> Unexpected use of undefined

**`app/(os-layout)/_components/TasksCenter.tsx:127`**

```
const predecessor = beforeIndex > 0 ? orderedSiblings[beforeIndex - 1] : undefined;
```

> Unexpected use of undefined

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

