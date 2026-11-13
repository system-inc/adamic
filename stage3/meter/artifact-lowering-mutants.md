# Artifact lowering mutant evidence (#sx8wdg7)

The meter's per-root column reads Lowering again. An ordinary root's cell is
production `lower.Lower` over the root's whole program (stage3/census/tool/main.go),
so it already includes every post-lowering pass. A root-order artifact root's
cell comes from the entry census instead, which lowers bodies and records the
whole-program passes as omitted (pass_ledger.py), so it now counts only through
`artifact_lowers_whole`: no body stop, no scoped tail stop, and every program
unit's whole-program passes measured. Today the census never measures them, so
no artifact root counts; the body-only result stays visible in the root-order
table and in the latent coverage lines.

Python cases are named by class and case only.

| Mutant | Line changed | Case it turns red | Expected result |
| --- | --- | --- | --- |
| Count artifacts by body, in the report | In `report`, replace `row['lowering'] = artifact_lowers_whole(artifact)` with `row['lowering'] = artifact_lowers_body(artifact)`. | `RootOrderTests.test_artifact_body_only_does_not_lower_sx8wdg7` | Fail: the root row reads lowering and the total is 2, against 1. Held in the suite by `RootOrderTests.test_artifact_body_count_mutant_is_red_sx8wdg7`, which patches the rule back and asserts the credit returns. |
| Count artifacts by body, in the root-order-only path | In `__main__`, replace `artifact_lowers_whole(row)` with `artifact_lowers_body(row)` in the totals. | none in the suite; that path writes root-order.json for a hand run and has no case | Survives. Named here so it isn't mistaken for covered. |
| Ignore the tail stop | In `artifact_lowers_whole`, drop `and not artifact.get('entry_tail_stops')`. | none alone: the same fixture has no measured passes | Survives alone, by design: `entry_whole_program_passes_measured` is false there too. With both terms dropped, `RootOrderTests.test_artifact_tail_stop_withholds_lowering_loom03dg` fails. |
| Keep the old label | In `root_order_table`, restore `Body lowering, {label}`. | `ReportTests.test_pair_preserves_area_fields_and_prints_four_numbers_first`, `RootOrderTests.test_daily_table_keeps_artifacts_beside_counts`, `RootOrderTests.test_artifact_body_only_does_not_lower_sx8wdg7` | Fail: each expects a line starting `Lowering, `. |
