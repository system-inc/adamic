# JSON values evidence

Implementation commit: c9d79bf83dde67bb31c0e1bf5f42fc13e394ca3b. Base main: 470e16f8315f7a83c5c7b7f6f2a77d8328db0a25.

Status: complete after the authorized precise full-reply harness repair.

- summary.json records the final measurements: 106,366 passing, one standing API acknowledgment failure, zero pending, with only api/typescript.d.ts moving.
- contracts.json records idempotence, no-write refusals, zero added casts, identical emitted JavaScript for the three harness declaration flows, nine red mutants and exact API reconstruction.
- measurement.json records all six stock projects clean and the native diagnostic multiset unchanged at 112 before and after.
- oracle-api-match.json records byte-identical reconstructed, generated and oracle declarations, and absence of the config error baseline.
- full-apply-source-match.json records exact source equality between a fresh end-to-end apply and the built, measured oracle tree.
- counterexample-response.json is the original real full reply. counterexample.json now proves that it is returned unchanged, including its string tag text.
- config-mutant-reuse.json proves that the unchanged compiler sources retain the existing red config-return mutant result in jv-config-mutant.json.
- full-reply-proofs.tar.gz contains final full apply/build logs, stock and native measurements, all nine mutant logs, oracle logs/report, reconstructed declaration, the prepared artifact hash manifest and command scripts.

proofs.tar.gz and harness-reach/ retain historical stopped-run evidence and the scratch reach measurement. Their stopped counts are superseded by summary.json and full-reply-proofs.tar.gz. The control and isolated config-return mutant remain applicable. No test expectation or baseline was accepted to hide a difference. The two completion-details baselines do not move in the final run.
