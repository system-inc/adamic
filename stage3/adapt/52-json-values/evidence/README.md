# JSON values evidence

Implementation commit: 0b5258a949ee4b55303b37ced1e95f06e984c326. Base main: 470e16f8315f7a83c5c7b7f6f2a77d8328db0a25.

Status: complete after correcting the unowned readConfigFile exemption.

- summary.json records 106,366 passing, one standing API acknowledgment failure, zero pending, and only api/typescript.d.ts moving. APISample_parseConfig.errors.txt is absent.
- contracts.json records idempotence, no-write refusals, zero added casts, unchanged harness type-flow JavaScript, thirteen red mutants and exact API reconstruction.
- measurement.json records six stock projects clean and native diagnostics unchanged at 112, with no additions or removals.
- oracle-api-match.json records byte-identical hand-applied sanctions, generated and oracle declarations. api-diff-counts.json records 135 diff lines (+19/-15).
- full-apply-source-match.json records exact source equality between the fresh end-to-end apply and the built oracle tree.
- harness-repair-unchanged.json proves the previously approved three-file harness repair stays byte-identical.
- jv-config-mutant.json records the freshly rebuilt parseConfigFileTextToJson return mutant and its targeted oracle: 5 passing, 1 failing, exactly APISample_parseConfig.errors.txt moves.
- own-pin-proofs.tar.gz contains final apply/build logs, both parallel Go test results, stock and native measurements, all thirteen mutant logs, the narrowed-return-only diagnostic probe, both fresh oracle reports and logs, reconstructed declaration, API diff, artifact hash manifests and oracle command script.

The single public-any exemption is parseConfigFileTextToJson.config. The Go test requires both its own ts.parseConfigFileTextToJson( call and its exact unchecked read; the maximum list length is one and zero is permitted. A second parallel test proves another function cannot borrow that call/read, and that the call or read alone is insufficient.

All ten design jsTyping diagnostics come from readConfigFile and now have real checks; there are no parseConfigFileTextToJson readers in jsTyping. Public explicit any falls from 37 to 22, with 21 outside the one-entry list. The excluded classes remain report-only.

proofs.tar.gz, full-reply-proofs.tar.gz and harness-reach/ retain historical measurements. Their earlier counts and two-entry exemption rationale are superseded by summary.json and own-pin-proofs.tar.gz. No baseline or test expectation was accepted to suppress a difference.
