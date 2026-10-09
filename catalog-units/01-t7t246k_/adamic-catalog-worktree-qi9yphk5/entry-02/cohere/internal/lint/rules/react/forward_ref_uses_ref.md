# `react/forward-ref-uses-ref`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require all forwardRef components include a ref parameter

## Why this recommendation

Incompatible with ESLint 10: the rule calls an API this version removed, so it throws rather than reports.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

