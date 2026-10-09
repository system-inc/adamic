# `max-lines`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **481** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a maximum number of lines per file

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

481 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:301`**

```
? allSearchResults.filter(function(task) {
```

> File has too many lines (1027). Maximum allowed is 300

**`app/(os-layout)/_components/drag/TasksDragContext.tsx:301`**

```
const sourceRegistration = rowRegistrationsReference.current.get(input.taskId);
```

> File has too many lines (663). Maximum allowed is 300

**`app/(os-layout)/_components/kingdom/OsKingdomActivity.tsx:301`**

```
/*
```

> File has too many lines (511). Maximum allowed is 300

**`app/(os-layout)/_components/kingdom/OsKingdomGraphCanvas.tsx:301`**

```
return () => cancelAnimationFrame(frame);
```

> File has too many lines (408). Maximum allowed is 300

**`app/(os-layout)/_components/kingdom/OsKingdomLayout.ts:301`**

```
* identical is returned as-is so React Flow skips it. Nodes with no fresh member
```

> File has too many lines (442). Maximum allowed is 300

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

