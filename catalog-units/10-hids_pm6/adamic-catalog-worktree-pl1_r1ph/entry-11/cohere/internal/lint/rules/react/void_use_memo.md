# `react-hooks/void-use-memo`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | none |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Validates that useMemos always return a value and that the result of the useMemo is used by the component/hook. See [`useMemo()` docs](https://react.dev/reference/react/useMemo) for more information.

## Why this recommendation

Stylistic, but the tree already satisfies it, so enabling costs nothing and prevents drift.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

