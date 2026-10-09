# `react/jsx-pascal-case`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce PascalCase for user-defined JSX components

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

