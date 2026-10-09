# `@typescript-eslint/parameter-properties`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **65** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require or disallow parameter properties in class constructors

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

65 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:460`**

```
constructor(public dynamicProperty: string) {}
```

> Property dynamicProperty should be declared as a class property

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.ts:11`**

```
private timeoutDuration: number,
```

> Property timeoutDuration should be declared as a class property

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.ts:12`**

```
private onFlush: (items: T[]) => void | Promise<void>,
```

> Property onFlush should be declared as a class property

**`libraries/structure/libraries/nexus/source/coordination/BatchingQueue.ts:13`**

```
private maximumBatchSize?: number,
```

> Property maximumBatchSize should be declared as a class property

**`libraries/structure/libraries/nexus/source/coordination/DeferredValue.ts:36`**

```
constructor(private readonly initializer: () => T) {}
```

> Property initializer should be declared as a class property

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

