# `react/jsx-closing-tag-location`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce closing tag location for multiline JSX

## Why this recommendation

Formatting, which Prettier already decides. Turning it on creates a second authority that can disagree.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

