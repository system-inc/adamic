# `no-array-constructor`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | none |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow `Array` constructors

## Why this recommendation

Misfiled as formatting; it is a correctness rule already covered by verify's own `no-array-constructor`, so the reason to decline is duplication rather than conflict with Prettier.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong No** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

