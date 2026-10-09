# `react/react-in-jsx-scope`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow missing React when using JSX

## Why this recommendation

Nothing to do with Prettier; it is obsolete under React 19's automatic JSX runtime and would be actively wrong if it ever fired, so decline it as obsolete rather than as a formatting conflict.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

