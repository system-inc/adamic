# `react/jsx-uses-react`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow React to be incorrectly marked as unused

## Why this recommendation

It is a no-unused-vars helper for the legacy JSX transform, not formatting; on React 19 with the automatic runtime it is simply inert, so the right verdict is a plain No.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

