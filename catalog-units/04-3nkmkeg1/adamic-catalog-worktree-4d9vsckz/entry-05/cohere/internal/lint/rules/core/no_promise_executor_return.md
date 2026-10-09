# `no-promise-executor-return`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **61** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow returning values from Promise executor functions

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

61 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/command-line/Readline.ts:12`**

```
rl.question(query, function(ans) {
```

> Return values from promise executor functions cannot be read

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:85`**

```
const timeout1 = new Promise((resolve) => setTimeout(resolve, 10));
```

> Return values from promise executor functions cannot be read

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:91`**

```
const timeout2 = new Promise((resolve) => setTimeout(resolve, 10));
```

> Return values from promise executor functions cannot be read

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:119`**

```
const timeout = new Promise((resolve) => setTimeout(resolve, 10));
```

> Return values from promise executor functions cannot be read

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:156`**

```
const timeout = new Promise((resolve) => setTimeout(resolve, 50));
```

> Return values from promise executor functions cannot be read

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

