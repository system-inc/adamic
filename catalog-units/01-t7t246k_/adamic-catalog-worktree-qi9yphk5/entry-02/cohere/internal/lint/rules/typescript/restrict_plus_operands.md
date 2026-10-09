# `@typescript-eslint/restrict-plus-operands`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | none |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Require both operands of addition to be the same type and be `bigint`, `number`, or `string`

## Why this recommendation

Catches a defect rather than a preference, and the tree is already clean, so it is a guardrail bought for nothing.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

