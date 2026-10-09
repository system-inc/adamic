# `no-empty-function`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **45** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow empty functions

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

45 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/coordination/BackoffTask.test.ts:115`**

```
const task = jest.fn().mockImplementation(() => new Promise(() => {})); // Never resolves
```

> Unexpected empty arrow function

**`libraries/structure/libraries/nexus/source/coordination/BackoffTask.test.ts:186`**

```
.mockImplementation(() => new Promise(() => {})); // Never resolves
```

> Unexpected empty arrow function

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.ts:12`**

```
this.resolveFunction = () => {};
```

> Unexpected empty arrow function

**`libraries/structure/libraries/nexus/source/coordination/PromiseBarrier.test.ts:65`**

```
const promise = Promise.reject(new Error('test error')).catch(() => {});
```

> Unexpected empty arrow function

**`libraries/structure/libraries/nexus/source/coordination/PromiseBarrier.test.ts:171`**

```
const promise1 = pendingPromise1.catch(() => {});
```

> Unexpected empty arrow function

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

