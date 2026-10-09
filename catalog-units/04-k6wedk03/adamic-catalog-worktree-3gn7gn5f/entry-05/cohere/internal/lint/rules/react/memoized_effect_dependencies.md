# `react-hooks/memoized-effect-dependencies`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | none |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Validates that effect dependencies are memoized

## Why this recommendation

Catches a defect rather than a preference, and the tree is already clean, so it is a guardrail bought for nothing.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

