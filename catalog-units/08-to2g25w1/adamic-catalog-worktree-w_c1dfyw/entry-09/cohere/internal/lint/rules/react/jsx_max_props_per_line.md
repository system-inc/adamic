# `react/jsx-max-props-per-line`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce maximum of props on a single line in JSX

## Why this recommendation

Formatting, which Prettier already decides. Turning it on creates a second authority that can disagree.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

