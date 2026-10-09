# `no-unused-expressions`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | none |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unused expressions

## Why this recommendation

Misfiled as a Prettier formatting concern; it is a correctness rule and verify already enforces `no-unused-expressions`, so the correct ground for declining is in-house duplication, not formatting authority.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

