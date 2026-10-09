Built fixture, meter and census case shards with hash-verified prepared executables.
Branch: codex/test-split-stage3-packages; fixture commit: 723acbcf233643cdada9ac594cba5cffe6f1eede.
173 fixtures pass; all owned shards are under 30 s; full tier and all Go units measured on four CPUs.
Fixture mismatch and meter/census dispatch mutants fail only their assigned shards; product and selector guards catch mutants.
Limits: apply and upstream lane exceed 30 s in other workers' territory; pre-existing probe failures remain; native fixture compilation stays in the hook.

Update, October 9: fixture tests now prepare missing products automatically.
[The self-preparation report](SELF-PREPARE-REPORT.md) supersedes the earlier
manual-preparation requirement and includes current-main cold compiler timings.


## Measurement context

Linux, Go 1.27.1, Node 24.19.0, clang 20.1.8. nproc=5;
cgroup cpu.max=400000 100000, exactly four CPU-seconds per second.
GOMAXPROCS=4 and fixture parallelism=4. Each measurement uses a fresh process,
-count=1 and ADAMIC_GATE_UNCACHED=1, with no successful observation reuse.
These are cold test observations after preparation, not a flushed Linux page
cache or an empty Go compilation cache. Test output always goes to files.
Go test binaries were compiled before measured invocations. Build-only rows
measure go test -run '^$' on the prepared Go action cache.

Setup timings: Go 0.097 s, Node 0.101 s, clang 0.575 s, Markdown dependencies
1.221 s, submodules 18.891 s, Go build 182.076 s, build cache warm 182.172 s,
done 182.198 s. GOPROXY was exported before setup. Setup exited zero and printed
/workspace/adamic-tools/env.sh, sourced in every build/test shell.
The first fixture measurement lacked @types/node and failed in 28.896 s.
After npm ci --prefix stage3/api and installing pinned TypeScript 6.0.3 for the
lane API parser, the original unchanged fixtures binary passed in 18.790 s.
The corrected measurement supersedes that failed dependency attempt. Go units
were measured again after preparation completed; optional Python integration
probes were repeated with actual CENSUS_BINARY, LATENT_CENSUS_BINARY and
ENTRY_CENSUS_BINARY products. Initial skips do not count as integration passes.
Initial replay measurements overlapped overlay preparation; the isolated repeat
measurements below supersede them.

## Scope and existing failures

The authoritative go list ./stage3/... inventory contains 14 packages: four
with tests and ten build-only packages. All their discovered test entry points
are below. Python discovery also covers lane, verdict, meter and census replay.
No standalone Python/Go tests were found under stage3/oracle; its run.py runs
the upstream suite assigned to another worker. The drivers' only Go package is
the parser probe, which has no tests. The tier was read with git show origin/devtools/fast-gate:cloud/fast-gate/run.py
at branch commit a1011e89c5d4ebc364417d2c1cba6cc99eee4557,
file blob 5697f869b81d7a9eb81741b03a44b71742e38e57. That file was not edited.
Its three exact invocations and independent fresh STAGE3_CACHE directories are
listed below. The full upstream lane and apply are measurements in other
workers' territory, not confirmation runs or changes to those suites.

Existing current-main failures are retained exactly:

- FullEntryCensusTests.test_rejected_entry_measures_clean_nested_body_and_catches_first_error_mutant:
  expected three Refused sites, observed zero. With real binaries this fails in
  0.278 s; the ordinary environment would skip it.
- Both census ReplayTests fail in setUpClass before their test methods run.
  The nested unit has status panic, rather than attempted. The actual diagnostic
  is latent state copy: unexported IR field argumentFacts. Repeat exits are 5
  because zero tests ran and setup failed. No assertion or expected value was changed.

All owned fixture, meter and census shards are below thirty seconds. Apply and
the full upstream lane exceed the budget and remain assigned to the other workers.
The meter's 36 methods and census's two replay methods are now dispatched by
ADAMIC_TEST_SHARD, including the probes already red on the base main revision
54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8. Replay's preparation builds were moved
behind build-probes.py; test observations still run fresh. No expected value,
assertion, compiler source or gate runner was changed. No aggregate gate or
whole Go package was run for confirmation.

The final followup tier, Go and probe measurements were run sequentially without
competing test processes. Earlier exploratory followup measurements that overlapped
preparation are excluded. The tier's npm install, apply, upstream build and suite
remain included exactly as the gate invokes them. They are not hash-fetched
upstream build products; that preparation belongs to the other workers.

## Before and after

The before column records the original independent measurements. Final Go units
were remeasured from freshly fetched binaries; the after column gives those actual
observations. Unchanged lane and verdict methods keep their original measurements.
All exact commands, exits, elapsed wall seconds and log paths are in the JSON
files in shard-evidence; compressed complete output logs are beside them.
The fixtures row's after value is the eight separately measured units below.
The dormant runner-guard hook skips normally; it is exercised by its guard test.
The Exit column records the original independent case exit; final shard exits
are in the meter/census table below (replay's old direct-unittest exit 5 becomes
shard-runner exit 1, retaining the same setup assertion).

| Unit | Before / unchanged observation (s) | After (s) | Exit |
| --- | ---: | --- | ---: |
| `stage3/adapt/75-optional-widening/coverage::build-only` | 0.071 | 0.061 | 0 |
| `stage3/census/latent/refusalrewrite::TestCompiler41231d51Shape` | 0.007 | 0.008 | 0 |
| `stage3/census/latent/refusalrewrite::TestMissingFunctionMutantFailsLoudly` | 0.002 | 0.002 | 0 |
| `stage3/census/latent/refusalrewrite::TestChangedVisitorMutantFailsLoudly` | 0.003 | 0.003 | 0 |
| `stage3/census/latent/refusalrewrite::TestVisitorsCollectContinueAndSkipDiagnosedBodies` | 0.249 | 0.204 | 0 |
| `stage3/census/latent/refusalrewrite/cmd::build-only` | 0.046 | 0.041 | 0 |
| `stage3/census/latent/replay::build-only` | 0.046 | 0.040 | 0 |
| `stage3/census/latent/replay/worker::build-only` | 0.035 | 0.031 | 0 |
| `stage3/census/latent/statecopy::TestUnknownMutableContainerFailsLoudly` | 0.045 | 0.047 | 0 |
| `stage3/census/latent/statecopy::TestUnknownForeignPointerFailsLoudly` | 0.049 | 0.044 | 0 |
| `stage3/census/latent/statementrewrite::TestMissingStatementFailsWithMethodName` | 0.048 | 0.042 | 0 |
| `stage3/census/latent/tool::build-only` | 0.027 | 0.029 | 0 |
| `stage3/census/predicates/tools/tool::build-only` | 0.031 | 0.032 | 0 |
| `stage3/census/tool::build-only` | 0.077 | 0.093 | 0 |
| `stage3/drivers/parser/probe::build-only` | 0.078 | 0.076 | 0 |
| `stage3/fixtures::TestFixtures` | 18.790 | eight fixture shards below; parent 12.752 | 0 |
| `stage3/fixtures::TestFixturePaths` | 0.008 | 0.008 | 0 |
| `stage3/fixtures::TestTransformedNodeRunnerGuardHook` | 0.008 | 0.007 | 0 |
| `stage3/fixtures::TestTransformedNodeRunnerGuard` | 0.030 | 0.030 | 0 |
| `stage3/ledger/checker-259::build-only` | 0.069 | 0.068 | 0 |
| `stage3/ledger/checker-259/rerun-2026-10-08::build-only` | 0.066 | 0.072 | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_sanctioned_observation_passes` | 0.508 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_unsanctioned_line` | 0.505 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_missing_sanctioned_line` | 0.508 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_changed_sanctioned_type` | 0.523 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_multiline_printer_form` | 0.525 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_proved_reference_changes_are_included` | 0.523 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_unproved_reference_change` | 0.481 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_extra_comment_line` | 0.482 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_source_pin` | 0.332 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_failed_install` | 0.535 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_unexpected_tests_exit` | 0.546 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_diff_must_describe_snapshots` | 0.593 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_method_conversion_is_not_a_property_union` | 0.506 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_manifest_includes_only_proved_property_handoffs` | 0.280 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_unknown_platform` | 0.317 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_unknown_platform_stops_runner_before_apply` | 0.312 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_measured_macos_counts` | 0.495 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_node_version_matches_report` | 0.516 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_wrong_passed_count` | 0.514 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_different_single_failure` | 0.535 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_wrong_baseline_path` | 0.532 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_failed_apply` | 0.577 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_wrong_oracle_exit` | 0.497 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_failed_build` | 0.563 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_zero_failure_is_not_the_expected_failure` | 0.529 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_pending_test` | 0.501 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_reporter_counts_must_agree` | 0.540 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_filtered_run` | 1.745 | unchanged | 0 |
| `stage3/lane/test_check.py::LandingLaneTests.test_missing_evidence` | 0.477 | unchanged | 0 |
| `stage3/lane/test_table.py::LaneTableTests.test_table_is_preserved_beside_report` | 0.121 | unchanged | 0 |
| `stage3/lane/test_table.py::LaneTableTests.test_no_arguments_creates_fresh_results` | 0.092 | unchanged | 0 |
| `stage3/lane/test_table.py::LaneTableTests.test_shell_entry_point_with_no_arguments` | 0.144 | unchanged | 0 |
| `stage3/verdict/test_verdict.py::Comparisons.test_independent_stream_mutants` | 0.104 | unchanged | 0 |
| `stage3/verdict/test_verdict.py::Comparisons.test_baseline_root_mapping_and_filename_mutant` | 0.115 | unchanged | 0 |
| `stage3/verdict/test_verdict.py::Comparisons.test_library_placeholder_census_mutant` | 0.102 | unchanged | 0 |
| `stage3/verdict/test_verdict.py::Comparisons.test_census_mutants` | 0.102 | unchanged | 0 |
| `stage3/verdict/test_verdict.py::Comparisons.test_eof_difference` | 0.098 | unchanged | 0 |
| `stage3/meter/report_test.py::ReportTests.test_checker_and_lowering_are_separate` | 0.068 | meter shard 5/8: 1.229 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_planted_diagnostic_changes_only_its_own_file` | 0.077 | meter shard 4/8: 1.225 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_global_and_external_diagnostics_are_not_attributed_to_root` | 0.069 | meter shard 2/8: 2.559 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_pair_preserves_area_fields_and_prints_four_numbers_first` | 0.081 | meter shard 1/8: 2.215 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_malformed_diagnostic_is_rejected` | 0.069 | meter shard 2/8: 2.559 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_missing_file_is_rejected` | 0.069 | meter shard 6/8: 1.562 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_duplicate_file_is_rejected` | 0.068 | meter shard 7/8: 1.606 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_missing_whole_program_is_rejected` | 0.071 | meter shard 3/8: 1.190 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_unknown_kind_is_rejected` | 0.071 | meter shard 4/8: 1.225 | 0 |
| `stage3/meter/report_test.py::ReportTests.test_whole_program_roots_are_checked` | 0.068 | meter shard 7/8: 1.606 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_planted_notyet_moves_only_its_reason_count` | 0.087 | meter shard 3/8: 1.190 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_unmatched_reason_is_owner_blank_and_in_unowned_first` | 0.072 | meter shard 5/8: 1.229 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_unowned_threshold_uses_either_tree_and_summarizes_tail` | 0.071 | meter shard 6/8: 1.562 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_exact_prefix_and_variance_owners` | 0.077 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_table_groups_same_reason_across_kinds` | 0.076 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_missing_source_is_rejected` | 0.071 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_measurement_label_is_required` | 0.074 | meter shard 3/8: 1.190 | 0 |
| `stage3/meter/report_test.py::LatentReportTests.test_top_ten_excludes_dependency_skips_and_orders_ties` | 0.074 | meter shard 6/8: 1.562 | 0 |
| `stage3/meter/report_test.py::LatentCensusTests.test_overlay_planted_notyet_moves_only_its_reason_count` | 0.147 | meter shard 1/8: 2.215 | 0 |
| `stage3/meter/report_test.py::CensusAttributionTests.test_type_error_in_dependency_changes_only_dependency_own_file` | 0.260 | meter shard 6/8: 1.562 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_imported_and_global_diagnostics_block_lowering` | 0.069 | meter shard 1/8: 2.215 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_clean_reach_ranks_unique_sites_with_owners` | 0.079 | meter shard 1/8: 2.215 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_wrong_root_is_rejected` | 0.071 | meter shard 2/8: 2.559 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_missing_reachable_file_record_is_rejected` | 0.070 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_blocked_stream_cannot_claim_lowering` | 0.071 | meter shard 4/8: 1.225 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_missing_entry_is_rejected` | 0.086 | meter shard 6/8: 1.562 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_checker_and_lowering_diagnostics_must_agree` | 0.068 | meter shard 3/8: 1.190 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_clean_reach_cannot_carry_rejected_program_label` | 0.069 | meter shard 1/8: 2.215 | 0 |
| `stage3/meter/entry_test.py::EntryReportTests.test_overlay_missing_hook_fails_loudly` | 0.079 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/entry_test.py::EntryCensusTests.test_reach_includes_dependency_outside_compiler_and_excludes_unused_file` | 0.285 | meter shard 7/8: 1.606 | 0 |
| `stage3/meter/entry_test.py::FullEntryCensusTests.test_rejected_entry_measures_clean_nested_body_and_catches_first_error_mutant` | 0.278 | meter shard 0/8: 2.951 | 1 |
| `stage3/meter/compiler_test.py::CompilerSelectionTests.test_default_uses_checkout_compiler_for_both_trees` | 0.788 | meter shard 1/8: 2.215 | 0 |
| `stage3/meter/compiler_test.py::CompilerSelectionTests.test_per_ref_builds_and_runs_each_pinned_compiler` | 1.405 | meter shard 2/8: 2.559 | 0 |
| `stage3/meter/compiler_test.py::CompilerSelectionTests.test_legacy_latent_mode_and_stale_full_claim` | 1.506 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/compiler_test.py::CompilerSelectionTests.test_invalid_latent_mode_fails_before_creating_a_run` | 0.132 | meter shard 0/8: 2.951 | 0 |
| `stage3/meter/compiler_test.py::CompilerSelectionTests.test_invalid_mode_fails_before_creating_a_run` | 0.139 | meter shard 1/8: 2.215 | 0 |
| `stage3/census/latent/replay/replay_test.py::ReplayTests.test_nested_findings_match_full_census_in_order` | 3.919 | census shard 0/2: 2.475 | 5 |
| `stage3/census/latent/replay/replay_test.py::ReplayTests.test_parent_sibling_boundary_fails_signature_assertion` | 3.885 | census shard 1/2: 2.450 | 5 |
| `stage3/lane::python discovery` | 7.315 | unchanged | 0 |
| `stage3/verdict::python discovery` | 0.097 | unchanged | 0 |
| `stage3/meter::python discovery` | 4.213 | eight meter shards below | 1 |
| `stage3/census/latent/replay::python discovery` | 3.922 | two census shards below | 5 |

| Fixture unit | Cases | Cold wall (s) | Exit |
| --- | ---: | ---: | ---: |
| `0/8` | 22 | 2.459 | 0 |
| `1/8` | 22 | 2.521 | 0 |
| `2/8` | 18 | 1.687 | 0 |
| `3/8` | 18 | 1.936 | 0 |
| `4/8` | 22 | 2.412 | 0 |
| `5/8` | 22 | 1.840 | 0 |
| `6/8` | 26 | 2.134 | 0 |
| `7/8` | 23 | 2.194 | 0 |

The new TestFixtureShardManifest unit passes in 0.013 s and lists all 173 cases
and their assigned shard. audit-shards.py independently enumerates status files
and computes SHA-256 assignment, compares the complete Go manifest, checks each
shard's actual named fixture execution events, rejects duplicates/overlap, and
requires their union to equal the independent census. cases.json includes every
fixture and its shard. Shard selection remains stable under -run filtering.

## Mutants and unchanged checks

The audit copies the fixtures and appends planted shard failure to the recorded
stdout of runner/03_return_true.a. The actual source is unchanged. It checks the
same recorded-Node byte comparison as the original runner. Exits across shards
0 through 7 are 0,0,0,0,0,0,0,1; shard 7 reports recorded Node byte comparison
failed at that fixture's node check. Every unit, including mutant units, remains
under thirty seconds. All Node, stage0 diagnostic, native, sanitizer, leak,
platform, path, provenance and safe-update checks remain intact.

Additional executed mutants:

- Corrupt copied product bytes: oracle build product hash mismatch.
- Set manifest action to all zeroes: invalid oracle build manifest.
- Remove the action manifest: oracle hook not prepared; no hook build occurs.
- Invalid selectors bad, 0/0, -1/8, 8/8 and 0/2/3 each fail at invalid ADAMIC_TEST_SHARD.

The valid copied-store control fetches exactly the original binary bytes without
creating any build log. Native compilation is still performed by the inherited
internal/oracle hook, counted within each shard's measured time. Only the Go
oracle hook was moved into hash-checked preparation. This does not establish
remote fetch of native binaries. The product store is local; no remote build
product service was supplied. Exporting that store is left to the gate build tier.

Verification: eight fixture shards passed, their independent union proof passed,
the planted failure was isolated, fixture path and transformed-runner guards
passed, hash/manifest/missing-product guards were killed, go vet ./stage3/fixtures
passed with empty output, and git diff --check passed. No new .a fixture or
counts row was added. No production compiler or another worker's suite was edited.

New Go entry point (included in the authoritative package inventory):

| Unit | Before | After (s) | Exit |
| --- | --- | ---: | ---: |
| `stage3/fixtures::TestFixtureShardManifest` | new | 0.010 | 0 |

## Exact Stage 3 tier invocations

These use the gate command arguments unchanged, fresh separate caches and the
shared bare TypeScript mirror. The lane tests' time includes npm ci. The source
branch was fetched solely to read the command list. Full commands, cache paths,
exits, elapsed seconds and complete compressed logs are in followup/tier.

| Gate unit | Cold before (s) | After | Exit |
| --- | ---: | --- | ---: |
| `stage3-apply-tests` | 100.963 | unchanged; other worker must split | 0 |
| `stage3-lane-tests` | 7.720 | unchanged; under budget | 0 |
| `stage3-lane` | 430.697 | unchanged; other worker must split | 0 |

The landing verdict requires exactly 106,366 passing tests and the one sanctioned
public API baseline mismatch, validated against the composed and sanctioned API.
The upstream subprocess itself exits 1 for that acknowledged mismatch; the lane
invocation exits 0. The exact passing/failing counts and API declaration counts
are retained in lane-results/report.json.

The commands are python3 stage3/test_apply.py; bash -c with the gate's exact
mkdir/copy/npm ci followed by python3 -m unittest test_check test_table from
stage3/lane; and bash stage3/lane/run.sh with a new results directory. The actual
shell string is recorded verbatim in tier/results.json.

## Meter and census shards

Prepare once with python3 stage3/meter/build-probes.py --prepare, then run
ADAMIC_TEST_SHARD=i/8 python3 stage3/meter/shards.py meter, or
ADAMIC_TEST_SHARD=i/2 python3 stage3/meter/shards.py census, directing
output to a log. Preparation also prepares the shared fixture oracle hook through its existing keyed helper.
Fetch verifies all eight listed product hashes, without building.
The case-ID SHA-256 manifest lists every case and its shard. Independent audit
selection proves 36 meter cases and two census cases exactly once in the union.
Input pairs within a comparison test stay together: both observations are needed
for its unchanged assertion. Census uses two units, one for each test case.

| Unit | Original aggregate (s) | After (s) | Cases | Exit | Planted additional failure |
| --- | ---: | ---: | ---: | ---: | --- |
| `meter 0/8` | 4.213 | 2.951 | 8 | 1 | none |
| `meter 1/8` | 4.213 | 2.215 | 7 | 0 | caught only here |
| `meter 2/8` | 4.213 | 2.559 | 4 | 0 | none |
| `meter 3/8` | 4.213 | 1.190 | 4 | 0 | none |
| `meter 4/8` | 4.213 | 1.225 | 3 | 0 | none |
| `meter 5/8` | 4.213 | 1.229 | 2 | 0 | none |
| `meter 6/8` | 4.213 | 1.562 | 5 | 0 | none |
| `meter 7/8` | 4.213 | 1.606 | 3 | 0 | none |
| `census 0/2` | 3.922 | 2.475 | 1 | 1 | caught only here |
| `census 1/2` | 3.922 | 2.450 | 1 | 1 | none |

Meter shard 0/8 retains FullEntryCensusTests' expected Refused=3, observed=0
failure. Census shards 0/2 and 1/2 retain the setup assertion: nested status
panic instead of attempted, caused by latent state copy: unexported IR field
argumentFacts. Logs contain the original failures. All enabled meter probes retain their assertions and are not skipped.

The audit appends a planted dispatch failure to the selected case after running
all unchanged assertions. This proves routing even when census setup is already
red; it does not claim the blocked census semantic mutant reached its assertion.
The audit compares the complete baseline failure/error sets with each mutant
run: only the targeted case's shard gets one added failure. Invalid selectors
bad, 0/0, -1/8, 8/8 and 0/2/3 fail with exit 2 before fetching products.
Product corruption, wrong action, missing-manifest, missing-product and invalid-hash
mutants fail the new helper guards without a rebuild. Duplicate/empty case
inventories and unknown planted cases fail their inventory guards; complete output and the executed mutant script
are retained in followup. Original census checks blocked by main's panic remain
unverified until the compiler owner fixes that pre-existing bug.

No new .a fixture or counts row was added. The content stores are local and
portable only with the same repository path and keyed inputs; remote transport
is not configured. Native fixture C compilation and the compiler-selection
witness builds remain in their respective measured units. Moving native compilation
out of the oracle hook requires editing internal/oracle/oracle_test.go, explicitly
excluded from this worker's territory; caching its successful observations instead
would remove checks, so that was not done. The gate's command
list remains unchanged; dispatching these shard commands is the gate owner's
integration step. No PR was opened and no main branch was pushed.
