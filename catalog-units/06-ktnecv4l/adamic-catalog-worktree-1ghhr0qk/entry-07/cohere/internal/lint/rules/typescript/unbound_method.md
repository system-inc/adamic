# `@typescript-eslint/unbound-method`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **18** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Enforce unbound methods are called with their expected scope

## Why this recommendation

It is deliberately turned off in VerifySettings.json and the examples file contains zero recorded violation sites for it, so the count of 18 cannot be corroborated against real code.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong Yes** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

18 in the tree. Showing the first few.

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

