# The four changes landed: before and after on main (in progress)

Asked: three runs back to back on one box, each built as the meter builds it plus the timing lines, over the same
adapted tree: main `95a7b63f` (before), main `51981922` (after, with d7dbbb13, b4eb424b, 609383f5 and 51981922),
and `51981922` with `07b24606` (#3hxh5jm) reverted.

## The meter's overlay refuses main 51981922 (observed)

`make_overlay.py` on `51981922` stops before any build ([overlay-refusal-51981922.log](overlay-refusal-51981922.log)):

```
panic: latent statement rewrite: lowering.statement: unsupported signature func(node *ast.Node) (result []ir.Statement, err error)
```

`statementrewrite` (stage3/census/latent/statementrewrite/main.go:44) accepts `lowering.statement` only as
`func(node *ast.Node) ([]ir.Statement, error)`. `f2cf1ebb` ("Map debug builds to source lines with budgeted
cache-safe skip-free tests", 2026-10-28 23:36 UTC) gave it named results, `(result []ir.Statement, err error)`, and
left `statementrewrite` as it was. Nothing under `stage3/census/latent/statementrewrite` changed since `95a7b63f`.
origin/main is `51981922` as of this push, so the daily meter's per-ref `build_census` (its `make_overlay.py`
line, under `set -e`) will stop at `latent-overlay` on this main. The same refusal hits the revert tree.
`95a7b63f` builds.

## The trees (observed)

At Performance's word, the fix is Correctness's `55c1c38f` (typescript/statementrewrite-named-results,
chg_0wdzk9qm: "statementrewrite takes lowering.statement by its types, not its names"), cherry-picked as a
throwaway onto both 51981922 trees, never pushed. A patch of my own, written first, was dropped unused.
`95a7b63f` needs nothing.

| label | tree | how |
|---|---|---|
| before | `95a7b63f` | main |
| after | `be99a01a` | `51981922` + `55c1c38f` cherry-picked (`-x`, clean) |
| revert | `589f337c` | `51981922` + `git revert --no-edit 07b24606` (`93245360`, clean, no fallback needed) + `55c1c38f` cherry-picked |

All three pin cohere `06af2e39` and TypeScript `d92d9bfe`, and each built as the meter builds it plus the timing
lines. `go test ./stage3/census/latent/statementrewrite/` is green on the after tree. The revert of `07b24606`
removes its lowering changes and its tests and fixtures (62 files, 593 lines). Runs follow, back to back in that order.
