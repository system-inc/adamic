# `@typescript-eslint/prefer-readonly`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **253** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Require private members to be marked as `readonly` if they're never modified outside of the constructor

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

253 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.ts:11`**

```
private timeoutDuration: number,
```

> Member 'timeoutDuration: number' is never reassigned; mark it as `readonly`

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.ts:12`**

```
private onFlush: (items: T[]) => void | Promise<void>,
```

> Member 'onFlush: (items: T[]) => void | Promise<void>' is never reassigned; mark it as `readonly`

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.ts:13`**

```
private maximumBatchSize?: number,
```

> Member 'maximumBatchSize?: number' is never reassigned; mark it as `readonly`

**`libraries/structure/libraries/nexus/source/coordination/CountdownLatch.ts:8`**

```
private promise: Promise<void>;
```

> Member 'promise' is never reassigned; mark it as `readonly`

**`libraries/structure/libraries/nexus/source/coordination/CountingSemaphore.ts:4`**

```
private queue: (() => void)[] = [];
```

> Member 'queue' is never reassigned; mark it as `readonly`

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

