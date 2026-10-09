# `react/jsx-newline`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require or prevent a new line after jsx elements and expressions.

## Why this recommendation

Formatting, which Prettier already decides. Turning it on creates a second authority that can disagree.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

