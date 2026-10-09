# `@typescript-eslint/prefer-find`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | none |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Enforce the use of Array.prototype.find() over Array.prototype.filter() followed by [0] when looking for a single result

## Why this recommendation

Stylistic, but the tree already satisfies it, so enabling costs nothing and prevents drift.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

