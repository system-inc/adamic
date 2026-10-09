# `react/jsx-one-expression-per-line`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require one JSX element per line

## Why this recommendation

Incompatible with ESLint 10: the rule calls an API this version removed, so it throws rather than reports.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

