# `@typescript-eslint/max-params`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **258** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a maximum number of parameters in function definitions

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

258 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/OsKingdomLayout.ts:81`**

```
function placeSubtree(
```

> Function 'placeSubtree' has too many parameters (6). Maximum allowed is 3

**`app/(os-layout)/_components/row/TaskListDayView.tsx:89`**

```
function buildDayBuckets(
```

> Function 'buildDayBuckets' has too many parameters (4). Maximum allowed is 3

**`app/(os-layout)/os/providers/OsProvidersConfluenceCanvas.tsx:29`**

```
function curvePath(startX: number, startY: number, endX: number, endY: number): string {
```

> Function 'curvePath' has too many parameters (4). Maximum allowed is 3

**`app/(os-layout)/os/sensation/_components/SensationPerceptionRibbon.tsx:124`**

```
export function expectedCadenceInMilliseconds(
```

> Function 'expectedCadenceInMilliseconds' has too many parameters (4). Maximum allowed is 3

**`app/(os-layout)/os/wisdom/_hooks/useWisdomRuling.ts:156`**

```
async function(
```

> Async function has too many parameters (4). Maximum allowed is 3

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

