# `@typescript-eslint/strict-void-return`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **85** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow passing a value-returning function in a position accepting a void function

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

85 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.test.ts:9`**

```
onFlushSpy = jest.fn();
```

> Value-returning function used in a context where a void function is expected

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:85`**

```
const timeout1 = new Promise((resolve) => setTimeout(resolve, 10));
```

> Value returned in a context where a void return is expected

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:91`**

```
const timeout2 = new Promise((resolve) => setTimeout(resolve, 10));
```

> Value returned in a context where a void return is expected

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:119`**

```
const timeout = new Promise((resolve) => setTimeout(resolve, 10));
```

> Value returned in a context where a void return is expected

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.test.ts:156`**

```
const timeout = new Promise((resolve) => setTimeout(resolve, 50));
```

> Value returned in a context where a void return is expected

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

