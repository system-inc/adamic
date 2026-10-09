# `no-iterator`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | none |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the use of the `__iterator__` property

## Why this recommendation

Catches a defect rather than a preference, and the tree is already clean, so it is a guardrail bought for nothing.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

