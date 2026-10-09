# `@typescript-eslint/no-redeclare`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **1** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow variable redeclaration

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

1 in the tree. 

**`libraries/structure/libraries/nexus/source/coordination/TrackedPromise.ts:34`**

```
export namespace TrackedPromiseType {
```

> 'TrackedPromiseType' is already defined

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

