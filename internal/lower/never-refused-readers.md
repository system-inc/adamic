# Never-refusal reader audit (#53r37a3)

All 19 admission-delta programs were checked individually. This table lists live promises, tests, generators and current evidence records. Path and old-text searches include repository text and compressed historical records. A dated observation is not silently relabeled as a new run.

| Program | Reader (file:line and test) | Records now |
| --- | --- | --- |
| `internal/conformance/testdata/refusals/any_value.a` | `internal/conformance/testdata/refusals/any_value.a:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/any_value.ts` | `internal/conformance/testdata/refusals/any_value.ts:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/array_length.a` | `internal/conformance/testdata/refusals/array_length.a:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/array_length.ts` | `internal/conformance/testdata/refusals/array_length.ts:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/eval.a` | `internal/conformance/testdata/refusals/eval.a:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/eval.ts` | `internal/conformance/testdata/refusals/eval.ts:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/function_call.a` | `internal/conformance/testdata/refusals/function_call.a:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/function_call.ts` | `internal/conformance/testdata/refusals/function_call.ts:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/new_function.a` | `internal/conformance/testdata/refusals/new_function.a:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/conformance/testdata/refusals/new_function.ts` | `internal/conformance/testdata/refusals/new_function.ts:2`<br>`internal/conformance/conformance_test.go:185 (TestEveryCaseKeepsTheSpecsPromise; Refused judge at :106)` | Refused; the header pins its construct-specific refusal phrase. No counts or separate stop record. |
| `internal/oracle/testdata/records_guarded/early_any.a` | `internal/oracle/records_guarded_test.go:70`<br>`internal/oracle/records_guarded_test.go:83 (TestRecordGuardedReadStopsEarlier)` | Refused only, new complete any rule, at 5:7; NotYet is no longer accepted for this case. |
| `internal/oracle/testdata/stdout_handle_view/n4_any_out.a` | `internal/oracle/testdata/stdout_handle_view/n4_any_out.a.stop:1`<br>`internal/oracle/stdout_handle_view_test.go:23 (TestStdoutHandleViews; stop read at :43)` | Full Refused diagnostic and fix at 2:7; already updated in the carried feature. No counts record. |
| `internal/oracle/testdata/stdout_handle_view/n4_any_out.ts` | `internal/oracle/testdata/stdout_handle_view/n4_any_out.ts.stop:1`<br>`internal/oracle/stdout_handle_view_test.go:23 (TestStdoutHandleViews; stop read at :43)` | Full Refused diagnostic and fix at 2:7; already updated in the carried feature. No counts record. |
| `stage3/adapt/99-correctness-axolotl-s/clear-mutant-length.ts` | `stage3/adapt/99-correctness-axolotl-s/direct-mutant.cjs:17`<br>`stage3/adapt/99-correctness-axolotl-s/direct-mutant-result.json:40`<br>`stage3/adapt/99-correctness-axolotl-s/clear-mutant-length.log:1`<br>`stage3/adapt/99-correctness-axolotl-s/README.md:129` | Refused length rule; generated classification, diagnostic and log refreshed. Splice controls still lower. Current generic clear is instantiated with number[]. |
| `stage3/adapt/99-correctness-axolotl-s/filterMutate-mutant-length.ts` | `stage3/adapt/99-correctness-axolotl-s/direct-mutant.cjs:17`<br>`stage3/adapt/99-correctness-axolotl-s/direct-mutant-result.json:20`<br>`stage3/adapt/99-correctness-axolotl-s/filterMutate-mutant-length.log:1`<br>`stage3/adapt/99-correctness-axolotl-s/README.md:129` | Refused length rule; generated classification, diagnostic and log refreshed. Splice controls still lower. Current generic clear is instantiated with number[]. |
| `stage3/census/repro/r12/main.a` | `stage3/census/repro/r12/main.a:1`<br>`internal/acheck/acheck_test.go:47 (TestEveryProgramOutsideAPackageKeepsItsPromise)`<br>`stage3/census/make_repros.py:19`<br>`stage3/census/data/repro_manifest.json:71`<br>`stage3/census/REPORT.md:217`<br>`stage3/census/write_report.py:55` | First-line Refused promise and generated action. Live diagnostic now 2:5 because of the new header; historical census observation at 1:5 is explicitly labeled historical. Numeric census data unchanged. |
| `stage3/drivers/scanner/evidence/land-area-next/witnesses/10-any.a` | `stage3/drivers/scanner/evidence/land-area-next/run-witnesses.py:41`<br>`stage3/drivers/scanner/evidence/land-area-next/witnesses/results.json:111`<br>`stage3/drivers/scanner/evidence/land-area-next/witnesses-final.log:10`<br>`stage3/drivers/scanner/evidence/land-area-next/witnesses/10-any.build.stderr:1`<br>`stage3/drivers/scanner/evidence/land-area-next/REPORT.md:1` | Refused any rule, actual stderr and compiler/source hashes; Node remains ok\n, exit 0. Historical initial log and ordered full-scanner walk are retained as historical, not current witness expectations. |
| `stage3/drivers/scanner/evidence/land-area-next/witnesses/14-any-callback.a` | `stage3/drivers/scanner/evidence/land-area-next/run-witnesses.py:41`<br>`stage3/drivers/scanner/evidence/land-area-next/witnesses/results.json:162`<br>`stage3/drivers/scanner/evidence/land-area-next/witnesses-final.log:14`<br>`stage3/drivers/scanner/evidence/land-area-next/witnesses/14-any-callback.build.stderr:1`<br>`stage3/drivers/scanner/evidence/land-area-next/REPORT.md:1` | Refused any rule, actual stderr and compiler/source hashes; Node remains ok\n, exit 0. Historical initial log and ordered full-scanner walk are retained as historical, not current witness expectations. |
| `stage3/fixtures/real/any-value/main.a` | `stage3/fixtures/real/status.json:141`<br>`stage3/fixtures/fixtures_test.go:242 (TestFixturesReal; full comparison at :493)`<br>`stage3/fixtures/real/any-value/README.md:8`<br>`stage3/fixtures/real/any-value/main.mutant.md:1` | Refused, complete diagnostic/fix at 4:36 and current_reason updated; census_count 30, Node bytes and provenance preserved. Old-row rollback fails exactly this fixture at fixtures_test.go:493. |

## Historical observations and implicit readers

`stage3/census/data/{rankings,top30,sites,summary}.json`, `data/{repros,policy-repros}.jsonl.gz`, latent/speculative census tables, and dated meter runs retain their measured compiler/source revisions. Their numeric counts are not recomputed by this diagnostic-only change. The current r12 promise and the regenerated report distinguish the new Refused rule from that old measurement.

`stage3/drivers/scanner/evidence/land-area-next/{stops.json,witnesses.log}` describe the original ordered scanner walk and initial probe run. They are not a claim that a full scanner walk was rerun here. The current two minimal-witness rows, final log and stderr files were regenerated with compiler/source hashes, and REPORT.md marks that boundary.

The scanner Python suite is owned by `stage3/lane.TestScannerScratchRunnerPython` (`stage3/lane/python_test.go:374`). It reads scanner driver/evidence contracts, not an exact current compilation golden for these two historical probes. The refreshed generator now asserts their Refused rule. The 99s live witness test is `direct-mutant.cjs`; its other contracts are `contract.cjs`, `clear-contract.cjs`, `compare.cjs` and the Node witness programs. No separate Go package lives in 99s.

The root census has no Go package that compiles repro/r12. Its historical `audit.py` checks the saved measurements and existence of repro files. The live whole-repository compilation reader is internal/acheck. Stage3/lane also owns the census pair and any-return suites. Whole stage3/census Go packages and stage3/lane are included in validation.

The ten conformance inputs are discovered by directory walking; the oracle inputs by their dedicated tests. Refused inputs have no emitted-IR .counts files to refresh. The two stdout_handle_view .stop records were already Refused in the carried feature. Owners remain adaptation 40 for any, adaptation 99s (Correctness) for array length, and #53r37a3 for eval/Function.

Other occurrences of the five old text fragments include unrelated NotYet constructs (for example other new-Identifier forms and non-array field writes), existing any cast/member refusals, and frozen measurements. Their rules are not widened or relabeled. The live 99s hard-coded NotYet assertion, the permissive oracle any-kind pin, r12 generator/promise and real status row are updated here.

## Every changed file

| File | Why |
| --- | --- |
| `cloud/fast-gate/never-refused-product-order.py` | Carried -7 four independent top-level mutant products; regenerated on main. |
| `cloud/fast-gate/product-order.json` | Carried -7 four independent top-level mutant products; regenerated on main. |
| `internal/conformance/testdata/refusals/any_value.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/any_value.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/any_value_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/any_value_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/array_length.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/array_length.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/array_length_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/array_length_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/eval.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/eval.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/eval_control_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/eval_control_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/eval_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/eval_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/function_call.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/function_call.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/function_control_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/function_control_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/length_control_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/length_control_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/new_function.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/new_function.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/new_function_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/new_function_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/tuple_length_control_fix.a` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/conformance/testdata/refusals/tuple_length_control_fix.ts` | Carried -7 Refused case, fix or admitted control in both source formats. |
| `internal/lower/any_relations_test.go` | Carried -7 Refused pins, controls, product tests or corrected expected-kind diagnostics. |
| `internal/lower/expression.go` | Carried -7 lowering-entry hooks and any fallback, preserving dormant generic bodies. |
| `internal/lower/library_node_callbacks_test.go` | Carried -7 Refused pins, controls, product tests or corrected expected-kind diagnostics. |
| `internal/lower/never-refused-readers.md` | New complete 19-program reader audit, historical/live boundary and change inventory. |
| `internal/lower/never-refused.mutants.md` | Carried -7 four NotYet classification mutants and independent producer names. |
| `internal/lower/never_refused.go` | Carried -7 four narrowly recognized Never refusals and actionable fixes. |
| `internal/lower/never_refused_test.go` | Carried -7 Refused pins, controls, product tests or corrected expected-kind diagnostics. |
| `internal/lower/object.go` | Carried -7 lowering-entry hooks and any fallback, preserving dormant generic bodies. |
| `internal/lower/statements.go` | Carried -7 lowering-entry hooks and any fallback, preserving dormant generic bodies. |
| `internal/lower/taste_representation_test.go` | Carried -7 Refused pins, controls, product tests or corrected expected-kind diagnostics. |
| `internal/oracle/records_guarded_test.go` | Pin early_any to Refused and its new rule; retain the other earlier-stop policy. |
| `internal/oracle/testdata/stdout_handle_view/n4_any_out.a.stop` | Carried -7 exact Refused any diagnostic. |
| `internal/oracle/testdata/stdout_handle_view/n4_any_out.ts.stop` | Carried -7 exact Refused any diagnostic. |
| `stage3/adapt/99-correctness-axolotl-s/README.md` | Explain current Refused evidence and retained historical attempts. |
| `stage3/adapt/99-correctness-axolotl-s/clear-mutant-length.log` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/adapt/99-correctness-axolotl-s/clear-mutant-length.ts` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/adapt/99-correctness-axolotl-s/clear-mutant-splice.log` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/adapt/99-correctness-axolotl-s/clear-mutant-splice.ts` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/adapt/99-correctness-axolotl-s/direct-mutant-result.json` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/adapt/99-correctness-axolotl-s/direct-mutant.cjs` | Require Refused in the two length mutants and describe current generic owners accurately. |
| `stage3/adapt/99-correctness-axolotl-s/filterMutate-mutant-length.log` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/adapt/99-correctness-axolotl-s/filterMutate-mutant-splice.log` | Regenerated by direct-mutant.cjs from current owners/current compiler, including both lowering splice controls. |
| `stage3/census/REPORT.md` | Render the current r12 promise separately from historical base observations; numeric census data unchanged. |
| `stage3/census/data/repro_manifest.json` | Generate and record r12's explicit Refused promise; preserve unrelated repros. |
| `stage3/census/make_repros.py` | Generate and record r12's explicit Refused promise; preserve unrelated repros. |
| `stage3/census/repro/r12/main.a` | Generate and record r12's explicit Refused promise; preserve unrelated repros. |
| `stage3/census/write_report.py` | Render the current r12 promise separately from historical base observations; numeric census data unchanged. |
| `stage3/drivers/scanner/evidence/land-area-next/REPORT.md` | Separate current minimal-witness refresh from the historical ordered scanner walk. |
| `stage3/drivers/scanner/evidence/land-area-next/run-witnesses.py` | Make witness refresh portable/selectable and require Refused for the two any probes. |
| `stage3/drivers/scanner/evidence/land-area-next/witnesses-final.log` | Regenerated only the two any witness observations, with compiler/source hashes and unchanged Node bytes. |
| `stage3/drivers/scanner/evidence/land-area-next/witnesses/10-any.build.stderr` | Regenerated only the two any witness observations, with compiler/source hashes and unchanged Node bytes. |
| `stage3/drivers/scanner/evidence/land-area-next/witnesses/14-any-callback.build.stderr` | Regenerated only the two any witness observations, with compiler/source hashes and unchanged Node bytes. |
| `stage3/drivers/scanner/evidence/land-area-next/witnesses/results.json` | Regenerated only the two any witness observations, with compiler/source hashes and unchanged Node bytes. |
| `stage3/fixtures/real/any-value/README.md` | Document the current Refused row. |
| `stage3/fixtures/real/any-value/main.mutant.md` | Measured one-row rollback mutant, catch line and independent passing controls. |
| `stage3/fixtures/real/status.json` | Capture the actual Refused status/current reason while preserving Node/provenance/counts. |
| `stage3/meter/owners.json` | Carried -7 adaptation 40/99s and task ownership for the new refusal reasons. |
