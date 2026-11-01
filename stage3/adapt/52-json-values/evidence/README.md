# JSON values composition evidence

Implementation commit: a99125a4e0808eeba5dce403d63182dd3293c4b7.

The current proof uses wave 6 at 63b44d41a8fe7e9ee707959b7153190e164148b0, 98-d at d26001bf42d4e6ec74ad41a3c1b96d8832431f88, 52 with the single-import fix, and main at 80c2d3282e44b527555d59690bb413cc883d2bf2. It is not a main-only measurement.

- composition.json records exact pins, scratch composition/control commits, all 59 candidate and 58 control completed adaptations, the native binary hash and the three scratch-only generated-manifest conflict resolutions. No adapter refuses in either real apply.
- import-composition.json records one named ts-module import, HashedSet after DocumentSpan, and the duplicate-import mutant's exact 99f refusal (2 !== 1) without writes.
- summary.json records both full oracles: 106,366 passing, one standing API acknowledgment failure, zero pending, with only api/typescript.d.ts moving. No config-error or completion-details baseline moves.
- measurement.json records six stock projects clean and native diagnostics unchanged at 8, with no additions or removals.
- contracts.json records idempotence, atomic no-write refusals, zero added casts, one module import, unchanged harness declaration-flow JavaScript, thirteen red check mutants and exact reconstruction from 52's 16 sanction rows.
- oracle-api-match.json records byte-identical 52 reconstruction, built and oracle declarations, SHA-256 9868791a1b24cb29cacc482db323ca18ab94a7d32ec7c42b7ba9b958d09e474b.
- generator-match.json distinguishes that strict 52 match from the full raw generator's inherited 97d compact-member layout mismatch. Running the exact upstream formatter does not remove that mismatch. The lane's actual normalized declaration changes agree, retaining every type-bearing AST child. No generator, 97d row or baseline was changed.
- counterexample.json records the real full completion reply returned by identity with its original string tag text.
- jv-config-mutant.json records a fresh composed parse-return mutant: successful normal build, then 5 passing, 1 failing, exactly APISample_parseConfig.errors.txt moves.
- wave6-composition-proofs.tar.gz contains the exact composition patch, input pins, complete apply/build logs, native streams and six stock projects, both full oracle reports/logs plus the targeted mutant, all thirteen check mutants, the duplicate-import mutant/script, the raw generator discrepancy and formatter experiment, declarations, Go tests, artifact manifests and command scripts.

The single exemption remains parseConfigFileTextToJson.config. Public any remains 37 before and 22 after, with 21 outside that list. The implementation branch contains only 52 changes and its earlier test file; the scratch composition merge and its generated-manifest resolutions are not pushed there.

The older archives proofs.tar.gz, full-reply-proofs.tar.gz, own-pin-proofs.tar.gz and harness-reach/ retain historical main-only and stopped measurements. Their native counts and generator context are superseded by this composition proof. No test expectation or baseline was accepted.
