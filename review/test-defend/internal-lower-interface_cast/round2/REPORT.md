# Two-row defense, round 2

Starting origin/main: 76c59c81e8617cea1892a01841494895927a712c. nproc: 5. Warm env.sh worked; setup skipped. npm ci stage3/api: 1.201 seconds. Baseline green, wall 68.727 seconds. All 275 discovered top-level tests completed in each matrix. Two rows skipped: TestOriginalCycleLedger and TestOptionalWideningCensus. Uniqueness means the default enabled package suite; those skipped inventories remain unknown. No run exceeded the binary budget, and no panic interrupted discovery.

CODE UNDER TEST: Adamic's Go Lower implementation, especially checked tagged-interface casts/view field access and custom iterator array spread representation validation. ORACLE: NeedsNoFlag executes original TypeScript on Node and compares lowered JavaScript stdout/stderr/exit through the existing agreement helper (external-run). IteratorGaps uses hand-written expectations that errors.As finds *NotYet (self). No oracle or test was changed.

Coverage: NeedsNoFlag has zero exclusive covered source lines versus DefaultTaggedInterfaceAdmission. That table's unused-bad-factory case includes the same inline positive call/payload read, plus an unused function. Three shared-line semantic attempts were made: required-field admission, allowed-tag membership, and discriminant selection. All are caught by the admission row too. This finite defense did not establish uniqueness, and is not a deletion recommendation.

IteratorGaps has 1420 exclusive covered source lines versus IteratorViewsCannotEraseReceivers, across its seven inputs. Its final custom iterator number spread into (number|undefined)[] reaches the outer arrayLiteral guard at object.go:237. Dropping that refusal admits the unsupported representation and makes only IteratorGaps fail at iteration_test.go:30. The previous inner collection guard was not the guard deciding this case. G01/G02 were planned and vet-validated but not run after G03 established the defense. Only executed attempts are in rows.json.

Every standalone diff applies to the starting tree and passed go vet ./internal/lower/. Mutants were selected from a switched production source in one build, each with a separate ADAMIC_BUILD_CACHE_DIR. Switched source was restored. No tests, oracle or harness changed. matrix.json contains every passed, failed and skipped row; raw JSON logs contain subcases. plan.json maps edits to starting-commit lines. coverage-differences.json records exclusive blocks and lines.

Name/assertion finding for the row not defended: NeedsNoFlag checks a positive cast under the inherited environment and Node agreement, but does not explicitly unset or toggle a tagged-interface feature flag. It demonstrates current default behavior rather than proving independence from every flag configuration. IteratorGaps checks only the NotYet category, not each refusal reason; this allowed earlier unrelated refusals to mask disabled inner guards, but G03 removes the actual outer guard and yields nil.

Brief feedback and time costs:
- The requested branch already existed from the previous wider defense. Work began detached at freshly fetched origin/main, then evidence was merged into the existing defense branch, preserving history without a force push.
- The audit's report is split across REPORT.md, rows.json, summary.json and other evidence. All report/code/oracle notes were read; audit.json and matrix.json are also retained.
- The package gained tests: 239 in the prior audit, 275 now. Current NeedsNoFlag assertions also execute Node agreement, stronger than the old audit's nil-error oracle. Historical conclusions cannot be transferred without rechecking bodies.
- Whole-package uniqueness cannot include inventory rows skipped behind optional corpus configuration. Those two are identified rather than reported as passing.
- No exclusive statement lines is not proof of duplication. The identical positive call in the admission table motivated three shared-line semantic mutations, which remain nonunique.
- An ambiguous optional-field anchor occurred twice, including legacyView. It was rejected before edits and narrowed to the intended unique guard; this was an execution correction, not a defect in the brief.
- The 90-second binary budget differs from wall time: G03 took 82.85 wall seconds including compilation; binary durations are in matrix.json. No narrowing was needed.

Timings: standalone vet checks and switch vet in runs.json, coverage runs in coverage-times.json, baseline setup/list timing in baseline-times.json. Matrix total wall seconds: 296.026. Full package runs, including all new enabled tests, were used. No repo-wide suite, native-only opt-in rerun, or skipped inventory corpus was covered. No main push and no PR.
