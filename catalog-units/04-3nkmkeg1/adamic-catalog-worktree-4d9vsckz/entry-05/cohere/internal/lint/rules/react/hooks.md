# `react-hooks/hooks`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **21** |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Validates the rules of hooks

## Why this recommendation

It validates the rules of hooks, which is a correctness rule and not stylistic, and at 21 sites the cleanup is bounded.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

21 in the tree. Showing the first few.

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

