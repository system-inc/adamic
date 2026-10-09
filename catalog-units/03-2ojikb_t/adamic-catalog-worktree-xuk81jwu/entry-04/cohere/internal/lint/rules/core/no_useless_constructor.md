# `no-useless-constructor`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **1** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unnecessary constructors

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

1 in the tree. 

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:24`**

```
constructor() {}
```

> Useless constructor

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

