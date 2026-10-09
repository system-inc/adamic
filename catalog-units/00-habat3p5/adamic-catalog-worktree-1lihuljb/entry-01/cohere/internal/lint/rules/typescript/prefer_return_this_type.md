# `@typescript-eslint/prefer-return-this-type`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | none |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce that `this` is used when only `this` type is returned

## Why this recommendation

Stylistic, but the tree already satisfies it, so enabling costs nothing and prevents drift.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

