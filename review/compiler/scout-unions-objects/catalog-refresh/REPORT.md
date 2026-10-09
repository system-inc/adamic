Built: refreshed catalog entry 08 context while preserving the exact planted narrowed-number-field bug.
Commits: patch ce90d1a0, on compiler/scout-unions-objects after 485dcd21.
Commands: entry 08 PASS, full catalog PASS across all 16 rows, integration lane checks PASS 2.8s.
Mutants: all 11 active catalog bugs caught with their unchanged required tests and diagnostics.
Not covered: five existing skipped catalog rows, the whole gate, or new compiler behavior.

Only verify/catalog/08-narrowed-number-field.patch and this review evidence change. No production code, catalog command, expected test, expected diagnostic, fixture or catalog status is changed. No new counts rows or Go tests are introduced. compiler/scout-unions-main remains at 884d907a.

The refreshed patch still removes exactly the same declared maybe-field read/unwrap block from internal/lower/object.go. It still permits a stale number-or-undefined field to be read as an unchecked number. The hunk header and surrounding context now match the checked boxed-scalar helper. mutation-identity.json proves the original and refreshed added/deleted lines are byte-for-byte identical. The catalog's historical made_against values are retained; this validation targets ce90d1a0.

Commands after sourcing /workspace/adamic-tools/env.sh, with all output saved to files:

- `git apply --check verify/catalog/08-narrowed-number-field.patch`: PASS.
- `timeout 900 verify/catalog/check.sh HEAD --entry 08 --jobs 1`: exit 0, applies-and-fails-as-recorded, total wall 265.882s. The uncached control passes; its cold trimpath compilation takes 229.074s while the reported test execution is 0.883s. It ran in a monitored background session and did not time out. The mutant exits 1 and the required fixture fails with the unchanged stdout differs diagnostic.
- `timeout 900 verify/catalog/check.sh HEAD --jobs 3`: exit 0, total wall 187.059s. All eleven active patches apply, their clean controls pass, and their required tests fail with the recorded diagnostic. Entries 10 and 13 through 16 remain skipped with the existing reasons. No other patch refresh is necessary.
- `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`: PASS, 2.8s, gofmt and tools on 20 Go files, t.Parallel on 4 test packages, vet 4 packages. Fetch is bounded at 120s, lane process at 300s, pipefail enabled. Final checks run again after committing the evidence.

Entry 08's mutant observation is unchanged in kind: the JavaScript backend prints interface field: undefined false, class field: undefined false, and alias: undefined NaN. Native instead prints interface field: NaN true, class field: NaN false, and alias: NaN NaN. Both complete with exit 0, so the oracle reports stdout differs; it also catches the missing inserted native panic. This is a runtime semantic failure, not a compilation failure. The original catalog test and required stdout diagnostic remain authoritative.

Every active catalog mutant and its required catcher:

| Entry | Bug | Required test(s) | Required diagnostic |
| --- | --- | --- | --- |
| 01 | shared-slice-append | TestNativeAgreesWithNode/internal/oracle/testdata/shared_slice_append.a | oracle_test.go:613: exit codes differ |
| 02 | liveness-throw | TestNativeAgreesWithNode/internal/oracle/testdata/throw_keeps_old_value.a | oracle_test.go:613: exit codes differ |
| 03 | defined-lent | TestNativeAgreesWithNode/internal/oracle/testdata/borrow_defined_lent.a | oracle_test.go:613: exit codes differ |
| 04 | borrowed-array-move | TestNativeAgreesWithNode/internal/oracle/testdata/borrow_element_super_move.a | oracle_test.go:613: exit codes differ |
| 05 | spread-method-reuse | TestNativeAgreesWithNode/internal/oracle/testdata/reuse_spread_method_alias.a | oracle_test.go:613: stdout differs |
| 06 | constructor-capture-region | TestNativeAgreesWithNode/internal/oracle/testdata/regions_constructor_capture.a | oracle_test.go:613: exit codes differ |
| 07 | borrowed-element-reads | TestNativeAgreesWithNode/internal/oracle/testdata/borrow_element_virtual_store.a | oracle_test.go:618: exit codes differ |
| 08 | narrowed-number-field | TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87_f1_field_narrowed.a | oracle_test.go:604: stdout differs |
| 09 | literal-undefined-field | TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a | oracle_test.go:613: stdout differs |
| 11 | refuse-definite-assignment | TestDefiniteAssignmentUsesReadiness | definite_assignment_test.go:42: 0 checked reads, want 1 |
| 12 | refuse-suppression-directives | TestSuppressionDirectivesAreRefused/@ts-expect-error/boolean.a//**_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/boolean.a//**_prose__*_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/boolean.a//*_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/boolean.a///_@ts-expect-error_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/call.a//**_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/call.a//**_prose__*_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/call.a//*_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/call.a///_@ts-expect-error_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/number.a//**_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/number.a//**_prose__*_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/number.a//*_@ts-expect-error_*/_<br>TestSuppressionDirectivesAreRefused/@ts-expect-error/number.a///_@ts-expect-error_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/boolean.a//**_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/boolean.a//**_prose__*_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/boolean.a//*_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/boolean.a///_@ts-ignore_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/call.a//**_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/call.a//**_prose__*_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/call.a//*_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/call.a///_@ts-ignore_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/number.a//**_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/number.a//**_prose__*_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/number.a//*_@ts-ignore_*/_<br>TestSuppressionDirectivesAreRefused/@ts-ignore/number.a///_@ts-ignore_ | suppression_directives_test.go:23: got <nil>, want directive refusal |

Raw controls, mutants, setup logs, per-run timing JSON and results manifests are in entry-08-evidence/ and all-entries-evidence/. The summary outputs are entry-08.log and all-entries.log. Catalog workers restored and removed their disposable worktrees. No test output was piped.

Toolchain preparation used GOPROXY='https://proxy.golang.org|direct' and `timeout 300 bash cloud/setup.sh`, then its environment file. Setup completed in 89.024s; Go ready 0.045s, Node ready 0.046s, markdown validation 0.012s and ready 0.124s, submodules 0.142s, clang ready 0.303s. The remaining build/cache timing lines are in setup.log. nproc=5, CPU quota=4. Native catalog observations use ASan/UBSan, strict C11 warnings, release builds and ADAMIC_GATE_UNCACHED=1.

This repairs step 17's integration catalog-apply failure without weakening entry 08 or planting a different bug. No PR is opened and only compiler/scout-unions-objects is pushed.
