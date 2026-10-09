# `react-hooks/todo`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **61** |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Unimplemented features

## Why this recommendation

Its own description is 'Unimplemented features' — it is an internal React Compiler diagnostic for constructs the compiler cannot yet handle, not a style rule, and its 61 hits are compiler bailouts that carry no action for us.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Strong No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

61 in the tree. Showing the first few.

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

