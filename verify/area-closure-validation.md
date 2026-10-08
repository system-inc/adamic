Closure merge: 273af2725507ff11fdbc50f7dba1ce21fb021582.
Updated area merge: b004da3dfbb59611c195a4f1bba8fd3f1301daf5 (2f312321), clean.
All closure lower/flow, convention, counts and selected oracle checks passed.
Mutants: dead boolean dispatch and 12 optional-widening refusals fail under source overlays; zero constituent count differs from Node in both backends.
Not covered: the full gate, full stage 1 packages and runtime emission of the twelve intentionally refused fs programs.

## Environment

GOFLAGS=-p=1 and GOMAXPROCS=2 for the valid runs. nproc=5 (CPU quota 4).
Successful setup: total 105.325s; Go 0.057s, Node 0.060s, markdown 0.181s,
submodules 0.251s, clang 0.418s, build 104.931s, defer 105.265s, warm 105.268s.
Earlier concurrent cold builds were killed by OOM and are not validation evidence.

## Blocker decisions

(a) TestParserHasNoUnusedOptionalMethodThunks expects three retained wrappers and
the counted convention. The area removed Statements_type's optional workaround.
Checker constituent recursion 09 is a positive Node comparison; counts refreshed.
(b) Scanner frame 01 retains the intentional var refusal. Twelve fs fixtures
retain optional-widening refusals and complete normalized diagnostic sidecars.
Taste 21 retains a temporary checked recursive field NotYet, explicitly not a
permanent refusal. Its source remains available under taste/refused.
(c) Parameter properties avoid property-declaration access; hand-built IR and
inherited field accesses retain checked storage metadata and readiness. The fs
options borrow remains narrow. Logical-never dispatch now excludes checker-never
whole expressions in known-dead boolean branches, retaining declared storage.
No production function-pointer casts were introduced.

## Exact checks

All output was redirected to /tmp/area-closure-*.log, never piped.

```sh
go test -v ./internal/oracle -run '^TestClosureMergeRefusals$' -count=1
go test -v ./internal/native -run '^(TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestClosureConvention.*)$' -count=1 -timeout 30m
go test ./internal/lower ./internal/flow
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
go test -v ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(arguments_length\.a|arguments_length_callbacks\.a|arguments_length_extended\.a|arguments_length_method\.a|arguments_length_no_reader\.a|arguments_length_reduce_left\.a|arguments_length_spread\.a|arguments_length_static\.a|arguments_length_static_constructor\.a|arguments_length_unrelated_type\.a|arguments_length_value\.a|arguments_length_value_count\.a|census_predicate_marker\.a|census_predicate_marker_live\.a|closure_convention_host24\.a|closure_convention_nested\.a|closure_convention_plain\.a|closure_convention_receiver_rest\.a|closure_convention_regexp_count\.a|host_array_predicate\.a|host_array_unknown_predicate\.a|host_empty_object_slots\.a|host_map_generic_union\.a|host_map_iterator_probe\.a|host_map_iterator_union\.a|host_map_union\.a|host_never_branches\.a|host_optional_intrinsic\.a|host_rest_union\.a|host_scanner_never\.a|host_undefined_value\.a|host_unknown_error_code\.a|host_void_method\.a|native\-arguments\-length\-value\.a|nested_array\.a|nested_assignment_return\.a|nested_callback_escaped\.a|nested_callback_late_capture\.a|nested_callback_levels\.a|nested_callback_recursive\.a|nested_captures\.a|nested_destructured\.a|nested_destructured_tdz\.a|nested_generic_capture\.a|nested_hoisting\.a|nested_minimal\.a|nested_mixed\.a|nested_mutual\.a|nested_owned_parameter\.a|nested_pattern_parameter\.a|nested_reduced_parameter_escaped\.a|nested_reduced_parameter_unreachable\.a|nested_reduced_parameter_unused\.a|nested_reference_identity\.a|nested_returned\.a|nested_tdz\.a|nested_tdz_write\.a|nested_three_levels\.a|nested_weak\.a|node_buffer_bom\.a|node_buffer_crypto\.a|node_buffer_encodings\.a|node_buffer_finalized\.a|node_buffer_input\.a|node_buffer_random\.a|node_buffer_utf16\.a|node_buffer_utf8\.a|node_buffer_writes\.a|node_fs_file_close\.a|node_fs_file_date\.a|node_fs_file_mkdir\.a|node_fs_file_write_buffer\.a|node_fs_file_write_file\.a|non_null\.a|non_null_array\.a|non_null_boolean\.a|non_null_capture\.a|non_null_catch\.a|non_null_definite\.a|non_null_definite_field\.a|non_null_definite_local\.a|non_null_field\.a|non_null_initialized\.a|non_null_lazy_default\.a|non_null_lazy_field\.a|non_null_lazy_initialized\.a|non_null_lazy_read\.a|non_null_lazy_static\.a|non_null_literal_assignment\.a|non_null_literal_return\.a|non_null_literal_statement\.a|non_null_map\.a|non_null_null\.a|non_null_static_initialized\.a|non_null_static_uninitialized\.a|non_null_uninitialized\.a|non_null_uninitialized_append\.a|non_null_uninitialized_capture\.a|non_null_uninitialized_catch\.a|non_null_uninitialized_const\.a|non_null_uninitialized_default\.a|non_null_uninitialized_exception\.a|non_null_uninitialized_field\.a|non_null_uninitialized_interface\.a|non_null_uninitialized_iteration\.a|non_null_uninitialized_loop\.a|non_null_uninitialized_map_entry\.a|non_null_uninitialized_optional\.a|non_null_uninitialized_spread\.a|non_null_union\.a|non_null_weak\.a|non_null_weak_freed\.a|omitted_boolean\.a|omitted_defaults\.a|omitted_methods\.a|omitted_number\.a|omitted_object\.a|omitted_reader\.a|omitted_reader_direct\.a|omitted_reader_override\.a|omitted_reader_string\.a|omitted_scanner\.a|omitted_scanner_explicit\.a|omitted_scanner_required\.a|omitted_string\.a|scanner_nested_overload\.a|taste_comma\.a|taste_labels\.a|taste_logical_assignment\.a|taste_optional_join\.a|taste_stage3_representations\.a|taste_truthiness\.a|taste_void\.a|unknown_narrowing\.a|unknown_narrowing_host\.a|arguments_length_refused|imported_const_cases|optional_widening_refused|regexp_replace|switch_case_declarations|taste_exports|taste_star)$' -count=1 -timeout 30m
go test -v ./internal/oracle -run '^(TestArgumentsLengthTypeScriptSource|TestArgumentsLengthWrongSlotMutant|TestClosedNestedConstituentMutant|TestClosureMergeRefusals|TestDefaultTaggedSourceViews|TestFieldReadinessRepresentation|TestImportedNonliteralConstCaseIsNotYet|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastOracle|TestInterfaceCastRuntimeMutants|TestInterfaceCastScalarTags|TestLazyInitializerIsNotEagerMutant|TestNarrowedFieldUsesSharedReadiness|TestNestedCallbackCarrierMutantIsCaught|TestNestedCycleRefusal|TestNestedRebindingCheckerRefusal|TestNestedReferenceIdentityMutant|TestNestedSiblingCycleMutantIsCaught|TestNodeFSFileAgreesWithNode|TestNodeFSFileMutants|TestNonNullWeakFreedNamesExpression|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestReadinessMutants|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRequiredViewFieldOperandOnce|TestRequiredViewFieldPrimitive|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestScannerNestedReferences|TestSwitchCaseDeadZoneStop|TestSwitchCaseDeclarationMutants|TestSwitchCaseOverloadIsNotYet|TestUninitializedIsNotNullishMutant|TestUnknownNarrowingMutants|TestViewFieldInheritedStaticReadiness)$' -count=1 -timeout 30m
go test -v ./internal/oracle -run '^TestNativeAgreesWithNode$/^stage3$/^(fixtures|drivers)$/^(nested-functions|taste|scanner)$/' -count=1 -timeout 30m
```

Outputs: refusals 5.039s; native convention 29.445s; lower 65.488s;
flow 289.849s; counts 137.365s; oracle batches 68.335s, 66.721s and 22.123s.
Counts changes are measured observations, not hand-edited predictions.
The updated area's lower/flow and catalog runs are recorded below when complete.

## Mutants

```sh
go test -overlay /tmp/area-closure-dead-boolean-overlay.json ./internal/lower -run '^TestCensusBooleanDeadBranch$' -count=1
go test -overlay /tmp/area-closure-optional-widening-overlay.json ./internal/oracle -run '^TestClosureMergeRefusals$/^internal$/^oracle$/^testdata$/^optional_widening_refused$/' -count=1
```

Both exit 1: the first restores a NotYet for type never; the second permits all
12 fs witnesses, failing their pinned refusal checks with got <nil>. Neither
failure comes from a build error. The retained Node-based runtime mutants ran
through the second oracle command above. TestClosedNestedConstituentMutant sets
only the implementation's returned count to zero: valid C, exit zero, no leaks
or sanitizer findings, and stdout disagrees with source Node in both backends.
TestOptionalMethodThunksMatchNode removes a required thunk and is caught by Node
output/exit comparison. Convention wrong-order and drop-count mutations are
rejected by clang's strict function-pointer signature checks.

The fs mutation table keeps all source cases but does not claim runtime evidence
for sources stopped by optional widening. Only the supported fs cases reach the
runtime mutants. No new source was copied from cohere.

## Updated area and catalog

The area merge changed no lower, native, flow or oracle source or fixtures.
The explicitly required `go test ./internal/lower ./internal/flow` rerun passed:
lower 79.725s, flow 289.231s. There was no additional oracle fixture selection
for that clean area merge.

```sh
bash verify/catalog/check.sh d772b5c990de4304de990e723835f19ebb613641 -jobs 1
```

The network run passed entries 01 and 02, then failed submodule setup after
GitHub rejected authentication. The repeat with pinned local Git submodule URL
rewrites completed in 239.488s: entries 01-03, 05-07 and 12 detect their faults;
04, 08, 09 and 11 report no-longer-applies. All five historical skipped entries
retain their reasons. The four patches are refreshed without changing compiler
production code; entry 11 targets the replacement readiness contract explicitly.

Refreshed catalog validation:

```sh
bash verify/catalog/check.sh 5c104a61e1156aa7a3cc14844df4f32d0fa76ee4 -jobs 1
```

Exit 0, 283.391s. All eleven active entries report
`applies-and-fails-as-recorded`; the five historical skips are not reassessed.
The full per-entry timing/status manifest is
catalog/logs/compiler-area-closure-results.json (external run evidence).
The catalog's exact named commands, expected failing subtests and diagnostics
remain in catalog.json. No compiler production code changed in this refresh.
