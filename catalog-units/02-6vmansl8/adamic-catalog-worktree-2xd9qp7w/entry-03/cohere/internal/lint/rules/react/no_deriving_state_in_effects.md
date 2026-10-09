# `react-hooks/no-deriving-state-in-effects`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | none |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Validates against deriving values from state in an effect

## Why this recommendation

Catches a defect rather than a preference, and the tree is already clean, so it is a guardrail bought for nothing.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

