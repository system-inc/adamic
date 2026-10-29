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

Next: a throwaway that lets `statementrewrite` accept the named results (renaming and wrapping work the same
either way), never committed into the meter, so the three runs can go ahead. Its diff and results follow.
