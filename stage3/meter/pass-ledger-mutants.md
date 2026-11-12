# Pass-ledger mutant evidence

These mutants exercise the ledger-only harness change. The positive assertions
remain unchanged; each mutant must turn its named case red. Python cases are
identified by class and case name only.

| Mutant | Line changed | Case it turns red | Expected result |
| --- | --- | --- | --- |
| Drop a required production step | In `stage3/census/latent/pass_ledger.go.txt`, remove `{Name: "checkOverloadResults"}` from `latentPostLoweringSteps`. | `TestLatentPassLedgerMatchesMainTailPb47` | Fail with `post-lowering step names/order drift`; the main tail still contains the missing call. The actual source mutation exited 1. |
| Rename a nested plan pass | In the guard's main-source copy, replace `lowering.applyProgramMetadata(plan)` with `lowering.unlistedPlanPass(plan)`. | `TestLatentPassLedgerMatchesMainTailPb47`'s tail comparison, held by `TestLatentPassLedgerNestedCallDriftPb47` | The comparison rejects the changed nested call even though the outer `if plan != nil` is unchanged. The negative-control case passes only when that rejection occurs. |
| Treat omissions as execution | In `stage3/census/latent/pass_ledger.py`, insert `return True` at the start of `whole_passes_measured`. | `PassLedgerReportTestsPb47.test_body_only_pass_ledger_report_line_pb47` | Fail: the expected line says `2 units lower without a body stop; 0 have their whole-program passes measured`, while the mutant prints 2 measured units. The actual source mutation exited 1. Held by `PassLedgerReportTestsPb47.test_omission_as_execution_report_mutant_is_red_pb47`. |
| Treat caller metadata as eligibility | In `stage3/census/pair/pair.py`, remove `instantiated_by` from the `stable` exclusion tuple, retaining the new `pass_ledger` exclusion. | `PairTests.test_deferred_caller_metadata_is_not_eligibility_abrq61p` | Error with `unit eligibility differs` for an otherwise unchanged eligible unit. Held by `PairTests.test_caller_metadata_eligibility_mutant_is_red_abrq61p`; exactly one error and no assertion failures are expected from the unchanged positive case. |

The restored implementation passed the whole latent Go tree, its real overlay,
the meter suite and the pairing suite. No post-lowering pass is executed by this
change. The real emission sample and validation totals remain under
`stage3/census/latent/evidence/pass-ledger/`.
