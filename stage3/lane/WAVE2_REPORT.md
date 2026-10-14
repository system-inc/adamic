Generated all 151 manifest Go shards, 17 apply shards and two lane unittest class units.
Branch typescript/stage3-lane-shards on b111200310fe3914a151c790148fef519e869274; actual final tip is recorded in delivery.
All 180 lane Go tests pass in fresh bounded processes; slowest 20.86s, 0 over 30s and zero over 60s.
Same-artifact union: 106367 test executions, 151 shards, exact identities/outcomes/multiplicities; missing-shard and coverage mutants caught.
No run.py, cloud/fast-gate, apply source/test or parent-branch files changed; no rebase onto the upcoming Builder key fix or full compiler gate claimed.

# Shard and test walls, slowest first

All 180 tests in stage3/lane ran individually with products prepared. This table
includes the seven existing product/key/manifest checks as well as all 173
newly generated, moved or added shard, Python and consistency units. Each fresh
process runs `go test ./stage3/lane -run '^NAME$' -count=1 -timeout 90s -json`,
with an outer 90-second process-group limit. Test wall includes product lookup,
all setup, verification, execution and result capture; process wall also includes
Go tooling startup. OS page cache is retained; Node compile cache is disabled
in shards. No test results are reused. Values are measured, not estimates.

| Go test | Test seconds | Process seconds | Over 30s | Exit |
|---|---:|---:|---|---:|
| TestShard_Stage3LaneFile010 | 20.86 | 21.585 | no | 0 |
| TestShard_Stage3LaneFile063 | 20.13 | 21.214 | no | 0 |
| TestShard_Stage3LaneFile007 | 19.98 | 20.897 | no | 0 |
| TestShard_Stage3LaneFile011 | 19.13 | 19.993 | no | 0 |
| TestShard_Stage3LaneFile001 | 18.95 | 19.733 | no | 0 |
| TestShard_Stage3LaneFile097 | 18.94 | 19.751 | no | 0 |
| TestShard_Stage3LaneFile026 | 18.82 | 19.951 | no | 0 |
| TestShard_Stage3LaneFile076 | 18.54 | 19.404 | no | 0 |
| TestShard_Stage3LaneFile125 | 18.53 | 19.405 | no | 0 |
| TestShard_Stage3LaneFile029 | 18.43 | 19.252 | no | 0 |
| TestShard_Stage3LaneFile048 | 18.41 | 19.148 | no | 0 |
| TestShard_Stage3LaneFile064 | 18.38 | 20.016 | no | 0 |
| TestShard_Stage3LaneFile025 | 18.22 | 19.046 | no | 0 |
| TestShard_Stage3LaneFile069 | 18.19 | 19.039 | no | 0 |
| TestShard_Stage3LaneFile098 | 18.19 | 19.050 | no | 0 |
| TestShard_Stage3LaneFile052 | 18.14 | 19.130 | no | 0 |
| TestShard_Stage3LaneFile096 | 18.11 | 18.908 | no | 0 |
| TestShard_Stage3LaneFile061 | 18.01 | 18.899 | no | 0 |
| TestShard_Stage3LaneFile110 | 17.93 | 18.870 | no | 0 |
| TestShard_Stage3LaneFile034 | 17.90 | 18.668 | no | 0 |
| TestShard_Stage3LaneFile045 | 17.89 | 18.744 | no | 0 |
| TestShard_Stage3LaneFile046 | 17.88 | 18.641 | no | 0 |
| TestShard_Stage3LaneFile018 | 17.87 | 18.854 | no | 0 |
| TestShard_Stage3LaneFile122 | 17.87 | 18.847 | no | 0 |
| TestShard_Stage3LaneFile085 | 17.81 | 18.706 | no | 0 |
| TestShard_Stage3LaneFile032 | 17.78 | 18.478 | no | 0 |
| TestShard_Stage3LaneFile044 | 17.78 | 18.756 | no | 0 |
| TestShard_Stage3LaneFile081 | 17.69 | 18.485 | no | 0 |
| TestShard_Stage3LaneFile077 | 17.68 | 18.597 | no | 0 |
| TestShard_Stage3LaneFile121 | 17.63 | 18.450 | no | 0 |
| TestShard_Stage3LaneFile128 | 17.62 | 18.434 | no | 0 |
| TestShard_Stage3LaneFile123 | 17.59 | 18.439 | no | 0 |
| TestShard_Stage3LaneFile030 | 17.57 | 18.448 | no | 0 |
| TestShard_Stage3LaneFile084 | 17.57 | 18.340 | no | 0 |
| TestShard_Stage3LaneFile038 | 17.54 | 18.395 | no | 0 |
| TestShard_Stage3LaneFile089 | 17.44 | 18.289 | no | 0 |
| TestShard_Stage3LaneSymlinkCase00 | 17.42 | 18.245 | no | 0 |
| TestShard_Stage3LaneFile021 | 17.41 | 18.193 | no | 0 |
| TestShard_Stage3LaneFile075 | 17.40 | 18.257 | no | 0 |
| TestShard_Stage3LaneSymlinkCase10 | 17.35 | 18.153 | no | 0 |
| TestShard_Stage3LaneFile117 | 17.30 | 18.198 | no | 0 |
| TestShard_Stage3LaneFile049 | 17.28 | 18.054 | no | 0 |
| TestShard_Stage3LaneFile100 | 17.28 | 18.063 | no | 0 |
| TestShard_Stage3LaneFile109 | 17.28 | 18.143 | no | 0 |
| TestShard_Stage3LaneFile002 | 17.27 | 18.088 | no | 0 |
| TestShard_Stage3LaneFile058 | 17.26 | 18.086 | no | 0 |
| TestShard_Stage3LaneFile062 | 17.26 | 18.090 | no | 0 |
| TestShard_Stage3LaneFile074 | 17.25 | 18.235 | no | 0 |
| TestShard_Stage3LaneFile054 | 17.24 | 18.368 | no | 0 |
| TestShard_Stage3LaneFile101 | 17.22 | 18.033 | no | 0 |
| TestShard_Stage3LaneFile036 | 17.17 | 18.289 | no | 0 |
| TestShard_Stage3LaneFile079 | 17.16 | 17.985 | no | 0 |
| TestShard_Stage3LaneFile037 | 17.09 | 17.898 | no | 0 |
| TestShard_Stage3LaneFile050 | 17.09 | 18.060 | no | 0 |
| TestShard_Stage3LaneSymlinkCase05 | 17.07 | 18.085 | no | 0 |
| TestShard_Stage3LaneFile059 | 17.05 | 18.294 | no | 0 |
| TestShard_Stage3LaneFile113 | 17.05 | 18.004 | no | 0 |
| TestShard_Stage3LaneFile023 | 17.01 | 18.176 | no | 0 |
| TestShard_Stage3LaneFile053 | 17.01 | 17.987 | no | 0 |
| TestShard_Stage3LaneFile106 | 17.01 | 17.855 | no | 0 |
| TestShard_Stage3LaneFile012 | 16.98 | 17.843 | no | 0 |
| TestShard_Stage3LaneFile131 | 16.96 | 18.136 | no | 0 |
| TestShard_Stage3LaneFile070 | 16.94 | 17.877 | no | 0 |
| TestShard_Stage3LaneFile116 | 16.94 | 17.830 | no | 0 |
| TestShard_Stage3LaneFile124 | 16.92 | 17.751 | no | 0 |
| TestShard_Stage3LaneFile104 | 16.86 | 17.697 | no | 0 |
| TestShard_Stage3LaneFile078 | 16.85 | 17.604 | no | 0 |
| TestShard_Stage3LaneFile095 | 16.82 | 17.661 | no | 0 |
| TestShard_Stage3LaneFile041 | 16.79 | 17.692 | no | 0 |
| TestShard_Stage3LaneFile080 | 16.79 | 17.664 | no | 0 |
| TestShard_Stage3LaneFile055 | 16.75 | 17.910 | no | 0 |
| TestShard_Stage3LaneFile114 | 16.74 | 17.698 | no | 0 |
| TestShard_Stage3LaneFile066 | 16.72 | 17.690 | no | 0 |
| TestShard_Stage3LaneFile092 | 16.71 | 17.504 | no | 0 |
| TestShard_Stage3LaneFile042 | 16.68 | 17.694 | no | 0 |
| TestShard_Stage3LaneSymlinkCase09 | 16.65 | 17.989 | no | 0 |
| TestShard_Stage3LaneFile132 | 16.63 | 17.479 | no | 0 |
| TestShard_Stage3LaneFile060 | 16.62 | 17.427 | no | 0 |
| TestShard_Stage3LaneFile118 | 16.61 | 17.536 | no | 0 |
| TestShard_Stage3LaneFile133 | 16.60 | 17.489 | no | 0 |
| TestShard_Stage3LaneFile107 | 16.59 | 17.534 | no | 0 |
| TestShard_Stage3LaneSymlinkCase04 | 16.59 | 17.422 | no | 0 |
| TestShard_Stage3LaneFile111 | 16.58 | 17.336 | no | 0 |
| TestShard_Stage3LaneFile056 | 16.55 | 17.396 | no | 0 |
| TestShard_Stage3LaneFile105 | 16.53 | 17.332 | no | 0 |
| TestShard_Stage3LaneFile009 | 16.52 | 17.372 | no | 0 |
| TestShard_Stage3LaneFile073 | 16.52 | 17.443 | no | 0 |
| TestShard_Stage3LaneFile057 | 16.49 | 17.386 | no | 0 |
| TestShard_Stage3LaneSymlinkCase12 | 16.45 | 17.430 | no | 0 |
| TestShard_Stage3LaneFile086 | 16.44 | 17.162 | no | 0 |
| TestShard_Stage3LaneFile071 | 16.42 | 17.393 | no | 0 |
| TestShard_Stage3LaneFile019 | 16.40 | 17.340 | no | 0 |
| TestShard_Stage3LaneFile099 | 16.40 | 17.146 | no | 0 |
| TestShard_Stage3LaneFile120 | 16.38 | 18.159 | no | 0 |
| TestShard_Stage3LaneFile017 | 16.37 | 17.188 | no | 0 |
| TestShard_Stage3LaneSymlinkCase11 | 16.37 | 17.129 | no | 0 |
| TestShard_Stage3LaneFile047 | 16.36 | 17.229 | no | 0 |
| TestShard_Stage3LaneFile065 | 16.36 | 17.352 | no | 0 |
| TestShard_Stage3LaneFile108 | 16.36 | 17.131 | no | 0 |
| TestShard_Stage3LaneFile091 | 16.35 | 17.489 | no | 0 |
| TestShard_Stage3LaneFile087 | 16.33 | 17.090 | no | 0 |
| TestShard_Stage3LaneFile115 | 16.32 | 17.102 | no | 0 |
| TestShard_Stage3LaneFile093 | 16.28 | 17.032 | no | 0 |
| TestShard_Stage3LaneFile102 | 16.26 | 17.032 | no | 0 |
| TestShard_Stage3LaneFile134 | 16.20 | 17.050 | no | 0 |
| TestShard_Stage3LaneFile033 | 16.19 | 17.051 | no | 0 |
| TestShard_Stage3LaneFile083 | 16.18 | 17.289 | no | 0 |
| TestShard_Stage3LaneFile103 | 16.16 | 16.935 | no | 0 |
| TestShard_Stage3LaneFile094 | 16.13 | 17.052 | no | 0 |
| TestShard_Stage3LaneFile130 | 16.13 | 16.892 | no | 0 |
| TestShard_Stage3LaneSymlinkCase08 | 16.11 | 17.085 | no | 0 |
| TestShard_Stage3LaneFile020 | 16.00 | 16.912 | no | 0 |
| TestShard_Stage3LaneFile126 | 16.00 | 16.791 | no | 0 |
| TestShard_Stage3LaneFile067 | 15.99 | 16.720 | no | 0 |
| TestShard_Stage3LaneFile068 | 15.99 | 17.001 | no | 0 |
| TestShard_Stage3LaneFile031 | 15.97 | 16.985 | no | 0 |
| TestShard_Stage3LaneFile082 | 15.96 | 16.797 | no | 0 |
| TestShard_Stage3LaneFile119 | 15.94 | 16.732 | no | 0 |
| TestShard_Stage3LaneFile027 | 15.93 | 16.786 | no | 0 |
| TestShard_Stage3LaneFile043 | 15.91 | 16.731 | no | 0 |
| TestShard_Stage3LaneFile051 | 15.81 | 16.584 | no | 0 |
| TestShard_Stage3LaneFile072 | 15.75 | 16.550 | no | 0 |
| TestShard_Stage3LaneFile040 | 15.70 | 17.377 | no | 0 |
| TestShard_Stage3LaneFile090 | 15.70 | 16.454 | no | 0 |
| TestShard_Stage3LaneFile127 | 15.70 | 16.582 | no | 0 |
| TestShard_Stage3LaneFile039 | 15.62 | 16.475 | no | 0 |
| TestShard_Stage3LaneFile112 | 15.51 | 16.329 | no | 0 |
| TestShard_Stage3LaneFile028 | 15.49 | 16.397 | no | 0 |
| TestShard_Stage3LaneFile129 | 15.47 | 16.280 | no | 0 |
| TestShard_Stage3LaneFile035 | 15.46 | 16.177 | no | 0 |
| TestShard_Stage3LaneFile088 | 15.43 | 16.231 | no | 0 |
| TestShard_Stage3LaneFile024 | 15.31 | 16.092 | no | 0 |
| TestShard_Stage3LaneFile015 | 15.22 | 15.993 | no | 0 |
| TestShard_Stage3LaneFile022 | 15.17 | 16.078 | no | 0 |
| TestShard_Stage3LaneFile014 | 15.16 | 16.362 | no | 0 |
| TestShard_Stage3LaneFile008 | 15.05 | 16.274 | no | 0 |
| TestShard_Stage3LaneSymlinkCase06 | 14.95 | 15.677 | no | 0 |
| TestShard_Stage3LaneFile006 | 14.90 | 15.630 | no | 0 |
| TestShard_Stage3LaneFile013 | 14.75 | 15.635 | no | 0 |
| TestShard_Stage3LaneFile016 | 14.53 | 16.677 | no | 0 |
| TestShard_Stage3LaneSymlinkCase07 | 14.52 | 15.427 | no | 0 |
| TestShard_Stage3LaneFile005 | 14.51 | 15.227 | no | 0 |
| TestStage3ApplyShard00 | 13.36 | 14.461 | no | 0 |
| TestShard_Stage3LaneSymlinkCase14 | 13.28 | 14.431 | no | 0 |
| TestShard_Stage3LaneSymlinkCase02 | 13.14 | 13.915 | no | 0 |
| TestShard_Stage3LaneByteComparisons | 12.88 | 13.787 | no | 0 |
| TestShard_Stage3LaneFile003 | 12.88 | 13.725 | no | 0 |
| TestShard_Stage3LaneSymlinkCase01 | 12.56 | 13.468 | no | 0 |
| TestShard_Stage3LaneSymlinkCase13 | 12.11 | 12.965 | no | 0 |
| TestShard_Stage3LaneSymlinkCase03 | 12.10 | 12.923 | no | 0 |
| TestShard_Stage3LaneSymlinkCase15 | 11.95 | 12.905 | no | 0 |
| TestShard_Stage3LaneFile004 | 11.85 | 12.657 | no | 0 |
| TestStage3LaneLandingLaneTests | 9.53 | 10.449 | no | 0 |
| TestStage3LaneHarnessTracksInputs | 3.06 | 3.794 | no | 0 |
| TestStage3LaneCompilerTracksInputs | 2.94 | 3.839 | no | 0 |
| TestStage3LaneInstallTracksInputs | 2.79 | 3.587 | no | 0 |
| TestProduct_Stage3LaneCompiler | 0.47 | 1.276 | no | 0 |
| TestProduct_Stage3LaneHarness | 0.47 | 1.272 | no | 0 |
| TestStage3ApplyShard06 | 0.44 | 1.320 | no | 0 |
| TestStage3ApplyShard16 | 0.42 | 1.371 | no | 0 |
| TestStage3ApplyShard01 | 0.39 | 1.276 | no | 0 |
| TestStage3ApplyShard04 | 0.39 | 1.243 | no | 0 |
| TestStage3ApplyShard03 | 0.38 | 1.421 | no | 0 |
| TestStage3ApplyShard07 | 0.38 | 1.272 | no | 0 |
| TestProduct_Stage3LaneInstall | 0.37 | 1.272 | no | 0 |
| TestStage3ApplyShard11 | 0.37 | 1.220 | no | 0 |
| TestStage3ApplyShard13 | 0.37 | 1.130 | no | 0 |
| TestStage3ApplyShard02 | 0.36 | 1.280 | no | 0 |
| TestStage3ApplyShard05 | 0.36 | 1.322 | no | 0 |
| TestStage3ApplyShard08 | 0.36 | 1.318 | no | 0 |
| TestStage3ApplyShard10 | 0.36 | 1.127 | no | 0 |
| TestStage3ApplyShard09 | 0.35 | 2.071 | no | 0 |
| TestStage3ApplyShard12 | 0.35 | 1.173 | no | 0 |
| TestStage3ApplyShard15 | 0.35 | 1.272 | no | 0 |
| TestStage3ApplyShard14 | 0.32 | 1.121 | no | 0 |
| TestStage3LaneLaneTableTests | 0.15 | 1.067 | no | 0 |
| TestPythonUnitsMatchCommands | 0.12 | 0.868 | no | 0 |
| TestUnionIdentityOutcomes | 0.08 | 0.926 | no | 0 |
| TestShardsMatchManifest | 0.04 | 0.922 | no | 0 |
| TestShardManifestRejectsMissingAndEmpty | 0.00 | 0.817 | no | 0 |

Coordinator: `timeout 7200 python3 -B stage3/lane/measure_go_units.py /workspace/scratch/hefhych-wave2/measure`,
then --resume after the apply archive setup correction described below. Each
child has its own hard limit. Its ledger contains every exact argv, measured
wall and exit, including the initial failed attempt. The final unit table takes
the last attempt per name; raw attempts remain in evidence. The run was checked
at intervals below one minute rather than waited on blindly.

# Executed union

Whole command: `bash stage3/lane/run.sh /workspace/scratch/hefhych-wave2/whole --artifact ARTIFACT`,
subprocess timeout580 with outer timeout600, exit0 in 347.226649s.
The unmodified runner used eight original workers, all runners, no filter and
--light=false. Original API sanction and all baseline checks passed.

`timeout 180 python3 -B stage3/lane/prove_union.py MEASUREMENTS WHOLE ARTIFACT OUTPUT`
runs merge_shards.py three times, each timeout90: complete union exit0, removal
of shard001 exit1, restored union exit0 with exactly the original output.

```json
{
  "status": "pass",
  "errors": [],
  "counts": {
    "pass": 106366,
    "fail": 1,
    "skip": 0
  },
  "whole_counts": {
    "pass": 106366,
    "fail": 1,
    "skip": 0
  },
  "units": 151,
  "tests": 106367,
  "artifact": "ea363c16f8a4a7ba016a77a62522ec2aef15569a93fe0f3df8f43c31f90cf04e",
  "failed_tests": [
    "unittests:: Public APIs for typescript.d.ts should be acknowledged when they change"
  ],
  "baseline_diffs": [
    "api/typescript.d.ts"
  ],
  "whole_verdict": "pass",
  "missing_tests": [],
  "extra_tests": []
}
```

The comparison uses (runner, file, full title, outcome) counters, retaining
multiplicity. It also preserves the original identity digest, counts, phase and
artifact identities, exact sanctioned failure, baseline path sets and raw diff
bytes, and rechecks the whole-run verdict. The previous names-only comparison
could not bind a status to an individual identity; it is strengthened here.
Drop-shard output names all 273 missing executions and no extras, rather than
only reporting a missing file. The captured Go outputs, whole task profile and
comparison outputs are committed under evidence/union/.

# Python-to-Go case map

The three old phase commands now have these executed homes:

| Old command | Old executed cases | Go home |
|---|---:|---|
| python3 stage3/test_apply.py | 25 | TestStage3ApplyShard00 through 16 |
| python3 -m unittest test_check test_table in stage3/lane | 40 | TestStage3LaneLandingLaneTests and TestStage3LaneLaneTableTests |
| stage3/lane/run.sh | 106367 | 151 TestShard_ units, proved by the same-artifact whole union above |

The historical SHARDS_REPORT.md describes 17 discovered apply cases. This tree
has 25; its existing modulo partition is used with 17 shards, without dropping
new cases. The bridge loads the original unittest suites, records actual starts
and outcomes, and every Go wrapper checks the exact observed case set. The
inventory check also compares Go AST wrapper names. No Python case is rewritten.
A second proof executes the native Python CLIs, with -v added only to expose
case identities for the lane command, and compares those observed cases with
all Go records. It reports 65 equal Python executions.

| Go unit | Exact Python cases | Test seconds |
|---|---|---:|
| TestStage3ApplyShard00 | ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean<br>ProductTests.test_restore_fetches_verified_product | 13.36 |
| TestStage3ApplyShard01 | ProductTests.test_build_with_pinned_packages_never_runs_npm<br>ProductTests.test_shards_partition_every_case | 0.39 |
| TestStage3ApplyShard02 | ProductTests.test_build_without_api_packages_installs_nothing<br>ProductTests.test_symlinked_directory_digests_as_its_link | 0.36 |
| TestStage3ApplyShard03 | ProductTests.test_corrupt_product_never_builds<br>ProductTests.test_typescript_source_is_the_pinned_gate_input | 0.38 |
| TestStage3ApplyShard04 | ProductTests.test_existing_output_is_refused<br>ProductTests.test_unconfigured_cold_cache_builds_without_remote_read | 0.39 |
| TestStage3ApplyShard05 | ProductTests.test_fetch_exposes_prepared_parser<br>ProductTests.test_unkeyed_parser_override_is_refused | 0.36 |
| TestStage3ApplyShard06 | ProductTests.test_input_key_covers_all_inputs<br>ProductTests.test_write_table_is_explicit | 0.44 |
| TestStage3ApplyShard07 | ProductTests.test_manifest_key_ignores_checkout_permissions<br>ProofShardTests.test_every_indexed_read_file_has_one_shard | 0.38 |
| TestStage3ApplyShard08 | ProductTests.test_manifest_mutant_is_rejected | 0.36 |
| TestStage3ApplyShard09 | ProductTests.test_missing_product_builds_and_reports | 0.35 |
| TestStage3ApplyShard10 | ProductTests.test_new_adaptation_directory_changes_hash | 0.36 |
| TestStage3ApplyShard11 | ProductTests.test_only_manifest_absence_is_a_cache_miss | 0.37 |
| TestStage3ApplyShard12 | ProductTests.test_packages_come_installed_from_the_lockfile | 0.35 |
| TestStage3ApplyShard13 | ProductTests.test_packages_digest_pins_every_byte | 0.37 |
| TestStage3ApplyShard14 | ProductTests.test_payload_manifest_mutant_is_rejected | 0.32 |
| TestStage3ApplyShard15 | ProductTests.test_payload_mutant_is_rejected | 0.35 |
| TestStage3ApplyShard16 | ProductTests.test_publish_hook_describes_local_payloads | 0.42 |
| TestStage3LaneLandingLaneTests | test_check.LandingLaneTests.test_changed_sanctioned_type<br>test_check.LandingLaneTests.test_diff_must_describe_snapshots<br>test_check.LandingLaneTests.test_diff_previews_are_first_40_lines_per_file<br>test_check.LandingLaneTests.test_different_single_failure<br>test_check.LandingLaneTests.test_error_text_drop_mutant_is_caught<br>test_check.LandingLaneTests.test_extra_comment_line<br>test_check.LandingLaneTests.test_failed_apply<br>test_check.LandingLaneTests.test_failed_build<br>test_check.LandingLaneTests.test_failed_install<br>test_check.LandingLaneTests.test_failure_cause_precedence_and_multiple_files<br>test_check.LandingLaneTests.test_failure_details_survive_missing_api_snapshot<br>test_check.LandingLaneTests.test_failure_text_and_diff_are_published<br>test_check.LandingLaneTests.test_filtered_run<br>test_check.LandingLaneTests.test_manifest_includes_only_proved_property_handoffs<br>test_check.LandingLaneTests.test_measured_macos_counts<br>test_check.LandingLaneTests.test_metadata_error_does_not_change_acceptance<br>test_check.LandingLaneTests.test_method_conversion_is_not_a_property_union<br>test_check.LandingLaneTests.test_missing_evidence<br>test_check.LandingLaneTests.test_missing_sanctioned_line<br>test_check.LandingLaneTests.test_multiline_printer_form<br>test_check.LandingLaneTests.test_node_version_matches_report<br>test_check.LandingLaneTests.test_observer_preserves_arguments_and_return_values<br>test_check.LandingLaneTests.test_original_mocha_event_retains_elapsed_and_stack<br>test_check.LandingLaneTests.test_pending_test<br>test_check.LandingLaneTests.test_proved_reference_changes_are_included<br>test_check.LandingLaneTests.test_reporter_counts_must_agree<br>test_check.LandingLaneTests.test_sanctioned_observation_passes<br>test_check.LandingLaneTests.test_source_pin<br>test_check.LandingLaneTests.test_unexpected_tests_exit<br>test_check.LandingLaneTests.test_unknown_platform<br>test_check.LandingLaneTests.test_unknown_platform_stops_runner_before_apply<br>test_check.LandingLaneTests.test_unproved_reference_change<br>test_check.LandingLaneTests.test_unsanctioned_line<br>test_check.LandingLaneTests.test_wrong_baseline_path<br>test_check.LandingLaneTests.test_wrong_oracle_exit<br>test_check.LandingLaneTests.test_wrong_passed_count<br>test_check.LandingLaneTests.test_zero_failure_is_not_the_expected_failure | 9.53 |
| TestStage3LaneLaneTableTests | test_table.LaneTableTests.test_no_arguments_creates_fresh_results<br>test_table.LaneTableTests.test_shell_entry_point_with_no_arguments<br>test_table.LaneTableTests.test_table_is_preserved_beside_report | 0.15 |

`TestPythonUnitsMatchCommands` compares discovery from the old commands with
python-units.json and all wrapper declarations. Each execution independently
checks successful outcomes and the exact runtime case IDs. The native CLI
comparison is in evidence/union/python-comparison.json.

# Generation and mutants

`python3 -B stage3/lane/generate_shards.py` generates shards_generated_test.go
from the complete shards.json. It writes no retyped argv or expectations:
functions bind a unit name, then the common reader retrieves its actual command
and counts. A full manifest fingerprint makes any command/count drift stale.
TestShardsMatchManifest checks canonical generated bytes and the complete Go
TestShard_ namespace, including tests added outside the generated file.

| Mutant | Catcher and observed result |
|---|---|
| Delete one generated test | TestShardsMatchManifest exits1, generated declarations differ |
| Add a unit to a private copy of shards.json | Actual TestShardsMatchManifest with that copy supplied to its generator exits1, generated declarations differ |
| Rename a generated test | TestShardsMatchManifest exits1, generated declarations differ |
| Add a TestShard_ function outside generated file | TestShardsMatchManifest exits1, AST manifest/test names differ |
| Delete apply Go wrapper16 | TestPythonUnitsMatchCommands exits1, Python Go unit names differ |
| Add a Python case to old test_table command | TestPythonUnitsMatchCommands exits1, old command case discovery differs |
| Drop an actually executed Python table case | TestStage3LaneLaneTableTests exits1, Python case identities differ despite a successful Python process |
| Ignore per-identity outcomes | TestUnionIdentityOutcomes exits1, swapped-outcome assertion fails |
| Ignore multiplicity | TestUnionIdentityOutcomes exits1, duplicated-identity assertion fails |
| Remove shard001 results from real union | merge_shards.py exits1, names missing shard and 273 missing test executions; restored union exits0 |
| Remove an observed case from a private copy of Python results | Native command comparison exits1 naming lane Python case union mismatch |

`timeout 180 python3 -B stage3/lane/prove_wave2.py` exits0; each mutant child
runs the actual Go catcher with -count1 -timeout90. The real files are restored
in finally. Additional CLI comparison commands and exact outcomes are retained
in evidence. No compiler/product is changed by a mutant.

# Environment, failures and validation

Toolchain command: `export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh`,
exit0; source /workspace/adamic-tools/env.sh. nproc5, cpu.max400000/100000,
a four-core quota. Setup timing lines:

```text
setup: node ready (0.022s)
setup: go ready (0.021s)
setup: submodules ready (0.069s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.011s
setup: markdown dependencies ready (0.072s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.173s)
setup: shared cache ready (1.011s)
setup: go build ready (23.322s)
setup: test binaries deferred (use --warm-tests) (23.418s)
setup: build cache warm (23.419s)
setup: build-flags commit=b111200310fe3914a151c790148fef519e869274 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.05 1/413 57865 load-after=1.34 0.32 0.16 1/418 58492
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (23.445s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.Dz6OjU
```

The final product was fetched, not rebuilt: compiler hit0.23s, harness hit0.11s,
top-level fetch test0.49s. Artifact identity is
ea363c16f8a4a7ba016a77a62522ec2aef15569a93fe0f3df8f43c31f90cf04e.
The plan was rebound after the merge runtime changed; no task/count partition
changed. All units use exactly that final product and runtime fingerprint.

The first whole run exited1 after93.175762s when a worker exited before the API
baseline existed. The post-failure cgroup counter was oom_kill=1; no pre-run snapshot was taken.
Memory pressure is an inference from that counter, the worker crash and5.7GB
of prepared products occupying tmpfs alongside eight upstream workers. The task's cache was
moved to disk in52.807248s, preserving its logical path with a directory symlink.
The retry passed with no additional OOM kill, unchanged worker count and checks.
The failed output and resource observations remain evidence; no green verdict
is claimed for it.

The first apply00 wrapper exited1 in5.92s because buildcache read mode could not
resolve a root-local untraced product from a detached checkout. The final wrapper
fetches product99 and supplies its verified products/ archive directory through
STAGE3_PRODUCT_STORE, retaining read mode so a miss cannot build. The unchanged
checkout-cleanliness case then passed. Its initial failed output is archived.

`timeout 90 python3 -B -m unittest discover -s stage3/lane -p 'test_*.py'`
exits0,74 fixtures in11.043s, including the updated exact-union diagnostics.
All180 package tests are run individually rather than forcing the entire
151-shard suite through one90s package deadline. Consistency and union fixtures
also passed before measurement. Changed package and reader checks, vet,
committed-tip lane checks, final path/trailer audit and push outputs are recorded
in the final delivery. No other Go package imports lane; applyproducts is an
unchanged dependency and its fetch path is exercised by all17 apply units.

Only stage3/lane/** is changed. No parent files, stage3/test_apply.py,
cloud/fast-gate or run.py change. The branch starts from the requested b1112003
without rebasing onto the forthcoming Builder key fix. No full compiler gate,
other platform, remote store latency or Loom submission is claimed.
