# `react-hooks/invariant`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1** |
| Plugin | `react-hooks` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Internal invariants

## Why this recommendation

Described as 'Internal invariants', it reports React Compiler internal assertion failures rather than user-fixable code, so a single hit is a compiler bug report and not a codebase defect.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

1 in the tree. 

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

