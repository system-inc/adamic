# A type alias as a sanctioned kind: its mutants (#w743639, adaptation 81)

`stage3/lane/generate-api.cjs` gains `TypeAliasDeclaration` in `sanctionKinds`, so an adaptation's
`api-sanctions.json` may name a type alias, whose owner type is the whole aliased type. Adamic allowed it narrowly on
#w743639: the kind only lets a declaration of that kind be named, and each sanction is still one exact line.

Run by J-zkkms2s (sealed, azure-proof, scratch commit carrying adaptation 81 and the change, never cut):

- **The change undone.** With `TypeAliasDeclaration` taken back out of `sanctionKinds`,
  `test_sanctioned_api_freshness.SanctionedApiFreshnessTests.test_lane_committed_sanction_matches_full_tree_generation`
  goes red: "canonical generator exit 1", generate-api refusing
  `adapt/81-readonly-option-values/api-sanctions.json row 1` as an invalid sanction. The file was put back byte for byte
  after.
- **A type alias changed with no sanction row still fails (Adamic's direction).** The tree with 81 applied, judged by
  `stage3-lane-byte-comparisons` against `sanctioned-api.json` and `shard-plan.json` as committed before the
  regeneration, is `fail` with one error:
  `unsanctioned API declaration: TypeAliasDeclaration:["ts","CompilerOptionsValue"]#0`.
- **The sanction without its change.** The tree without 81, judged against the regenerated sanction, is `fail` with
  one error: `missing sanctioned API declaration: TypeAliasDeclaration:["ts","CompilerOptionsValue"]#0`.
- **With both,** the unit passes with the manifest's counts, 277 sanctioned declarations.
