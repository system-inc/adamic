# `react-hooks/memo-dependencies`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **21** |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Validates that useMemo() and useCallback() specify comprehensive dependencies without extraneous values. See [`useMemo()` docs](https://react.dev/reference/react/useMemo) for more information.

## Why this recommendation

Missing useMemo/useCallback dependencies are stale-closure bugs, not a convention, and 21 sites is a bounded cleanup for a correctness guardrail.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

21 in the tree. Showing the first few.

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

