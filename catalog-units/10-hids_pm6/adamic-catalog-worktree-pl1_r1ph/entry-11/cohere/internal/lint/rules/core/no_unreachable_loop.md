# `no-unreachable-loop`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **1** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow loops with a body that allows only one iteration

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

1 in the tree. 

**`libraries/structure/code-quality/lint/rules/NetworkRequireHookRequestSuffixRule.ts:87`**

```
for(const [hookName] of networkServiceHooks) {
```

> Invalid loop. Its body allows only one iteration

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

