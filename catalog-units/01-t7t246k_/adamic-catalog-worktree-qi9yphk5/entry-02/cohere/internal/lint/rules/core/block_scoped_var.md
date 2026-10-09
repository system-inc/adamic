# `block-scoped-var`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | none |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce the use of variables within the scope they are defined

## Why this recommendation

Catches a defect rather than a preference, and the tree is already clean, so it is a guardrail bought for nothing.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

