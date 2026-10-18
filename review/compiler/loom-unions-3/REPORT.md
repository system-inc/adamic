Built: nullable string/object unions, checked stored scalar tags and optional boolean class fields recut on current main.
Commits: source bb8b389a3c3c567c29afe465e246c3ca9466f835; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; delivery SHA in final response.
Commands and outputs: exact bounded argv, exits and command seconds in checks.json; final positive checks passed.
Mutants: all 22 recorded source mutants and the recorded IR field-check mutant caught.
Not covered: full packages/full gate, WASI, general callable/object-kind unions, optional writes/delete and hidden never-array lowering.

No squash conflicts. Automatic merges in expression.go preserve main's newer infrastructure beside the source's lowering. counts.md was taken from main before exactly one Linux -update-counts run using house clang 20.1.8. Main 1f07ce2d has already removed the digest column again; digests key measurement products, and that current format is preserved.

Counts: sixteen new admitted fixture rows and one changed existing row, taste_stage3_representations.a: retains/releases 11/34 to 13/36; allocations, frees, peak and regions unchanged. A current-tree counterfactual overlay disables only the new checked string-field route and measures 11/34 again (counts-cause.log). The same overlay returns scout_boxed_scalar_fields.a to 95/150 from 114/169. Thus each changed row belongs to this lowering. counts-attribution.json and counts.patch hold every row. Only the admitted source fixtures and the known affected existing taste fixture were regenerated; no Mac table was retained.

The source placed six deliberate NotYet storage probes among runnable root fixtures. Main's strict flow corpus correctly failed to lower them. They are now under internal/oracle/testdata/scout_nullable_storage/, with unchanged bytes and all six existing oracle tests pointing to them. All six still run against Node and require the explicit lowering stop. No test was deleted and no flow/freshness test was edited. The touched taste subtests now call t.Parallel first. These are the only changes beyond the squash and regenerated counts.

Runtime +/- lines under internal/native/runtime are IDENTICAL to the source: both empty. The native union.go conversion change outside that directory is retained as written. runtime-comparison.json records the empty arrays.

All requested guards passed: TestEveryMapAProgramHoldsIsSorted (no new IR maps), TestRuntimeSetsAreEachUsed (no new runtime set), TestNativeAgreementRowsNameRegisteredPrograms (no new agreement call), TestEveryTestIsParallelOrSaysWhy (plus parallel nested taste leaves), and TestCallTargetReaders with -timeout 90s. No new compiler-consuming package or build product was added, so compiler-dependencies.json needs no entry. Products/tests add no build-directory or machine-path literals.

Focused checks cover admitted fixtures against source Node and both backends, native ASan/UBSan/leaks, the six storage refusals, checked field IR mutant, RegExp callback adapter, the two changed stage1 gap leaves, flow/freshness remainders, native consumer checks, guards and vet. Vet covers lower, native, oracle, stage1/cohere/selector and stage1/cohere/values. Every test uses -count=1 -timeout 90s -parallel 4 with GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1; subprocesses and drivers have hard limits. Output goes directly to logs. Individual leaf and parent seconds are in test-times.tsv/json; all are under 60 seconds. No split was needed.

The unit's shared setup completed with clang 20.1.8, Go 1.27.1, Node v24.19.0; /workspace/adamic-tools/env.sh was sourced for this branch. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). setup.log preserves timing lines: Node 0.019s, Go 0.021s, Markdown 0.058s, clang 0.129s, submodules 33.572s, shared cache 38.397s, Go build 324.920s, tests deferred 325.035s, cache warm 325.036s, done 325.089s. The initial TypeScript history fetch was interrupted; a depth-1 exact-pin fetch and normal blob materialization recovered it, as the first branch report records.

Assumption: deliberate NotYet probes belong in their own fixture directory, retaining their Node/refusal tests rather than weakening another package's runnable corpus. No cohere code was copied. No PR or Loom submission was made. One commit is pushed only to compiler/loom-unions-3, without force.

Commands and observed outcomes:

- counts-regeneration: exit 0, 18.687s.
- counts-table: exit 0, 3.076s.
- lower-focus: exit 0, 9.368s.
- oracle-focus: exit 0, 6.0s.
- fixtures: exit 0, 7.983s.
- regex-adapter: exit 0, 15.626s.
- stage1-gaps: exit 0, 15.439s.
- flow-consumers: exit 1, 20.496s.
- storage-after-move: exit 0, 10.933s.
- counts-cause: exit 0, 16.192s.
- flow-consumers: exit 0, 6.348s.
- fresh-consumers: exit 0, 7.405s.
- native-consumers: exit 0, 7.794s.
- guards-ir: exit 0, 13.291s.
- guards-runtime: exit 0, 4.541s.
- guards-registration: exit 0, 2.795s.
- guards-parallel: exit 0, 1.221s.
- readers: exit 0, 8.33s.
- vet: exit 0, 3.183s.
- mutant-null-box: exit 1, 10.679s.
- mutant-runtime-typeof: exit 1, 10.921s.
- mutant-lower-typeof: exit 1, 12.428s.
- mutant-nullable-admission: exit 1, 8.909s.
- mutant-equality-observation: exit 1, 16.496s.
- mutant-class-zero: exit 1, 15.442s.
- mutant-class-optional: exit 1, 9.315s.
- mutant-boxed-check: exit 1, 15.952s.
- mutant-boxed-observation: exit 1, 16.031s.
- mutant-object-boundary: exit 1, 15.958s.
- mutant-object-tag: exit 1, 15.761s.
- mutant-object-optional: exit 1, 15.938s.
- mutant-object-null-field: exit 1, 16.135s.
- mutant-string-check: exit 1, 6.05s.
- mutant-string-tag: exit 1, 15.56s.
- mutant-local-nullish: exit 1, 15.626s.
- mutant-regex-adapter: exit 1, 16.901s.
- mutant-shared-view: exit 1, 16.021s.
- mutant-lookup: exit 1, 15.726s.
- mutant-local-equality: exit 1, 15.719s.
- mutant-object-local-runtime: exit 1, 15.313s.
- mutant-object-local: exit 1, 5.641s.

Recorded source mutants and observed catchers:

- boxed-check: TestNativeAgreesWithNode_127/internal/oracle/testdata/scout_boxed_scalar_stale_field.a (2.76s), TestNativeAgreesWithNode_127 (0s).
- boxed-observation: TestNativeAgreesWithNode_088/internal/oracle/testdata/scout_boxed_scalar_fields.a (2.88s), TestNativeAgreesWithNode_088 (0s).
- class-optional: TestScoutClassBooleanKeepsUnsafeViewsRefused (0.06s).
- class-zero: TestNativeAgreesWithNode_042/internal/oracle/testdata/scout_class_boolean_fields.a (2.77s), TestNativeAgreesWithNode_042 (0s).
- equality-observation: TestNativeAgreesWithNode_035/internal/oracle/testdata/scout_nullable_strings.a (2.98s), TestNativeAgreesWithNode_035 (0s).
- local-equality: TestNativeAgreesWithNode_029/internal/oracle/testdata/scout_string_number_fields_locals.a (2.76s), TestNativeAgreesWithNode_029 (0s).
- local-nullish: TestNativeAgreesWithNode_029/internal/oracle/testdata/scout_string_number_fields_locals.a (3.03s), TestNativeAgreesWithNode_029 (0s).
- lookup: TestNullableObjectStorageLookup (2.49s).
- lower-typeof: TestNativeAgreesWithNode_091/internal/oracle/testdata/scout_tsc_nullish_content.a (0.03s), TestNativeAgreesWithNode_091 (0s).
- null-box: TestNativeAgreesWithNode_035/internal/oracle/testdata/scout_nullable_strings.a (2.62s), TestNativeAgreesWithNode_035 (0s).
- nullable-admission: TestNullableStringAdmission (0.03s).
- object-boundary: TestNativeAgreesWithNode_053/internal/oracle/testdata/scout_nullable_objects.a (2.86s), TestNativeAgreesWithNode_053 (0s).
- object-local-runtime: TestNativeAgreesWithNode_112/internal/oracle/testdata/scout_nullable_object_stale_local.a (2.69s), TestNativeAgreesWithNode_112 (0s).
- object-local: TestNativeAgreesWithNode_053/internal/oracle/testdata/scout_nullable_objects.a (2.66s), TestNativeAgreesWithNode_053 (0s).
- object-null-field: TestNativeAgreesWithNode_103/internal/oracle/testdata/scout_nullable_object_null_field.a (2.78s), TestNativeAgreesWithNode_103 (0s).
- object-optional: TestNativeAgreesWithNode_053/internal/oracle/testdata/scout_nullable_objects.a (2.84s), TestNativeAgreesWithNode_053 (0s).
- object-tag: TestNativeAgreesWithNode_119/internal/oracle/testdata/scout_nullable_object_stale_field.a (2.87s), TestNativeAgreesWithNode_119 (0s).
- regex-adapter: TestNativeAgreesWithNode_122/internal/oracle/testdata/regexp_replace/callback.a (0.06s), TestNativeAgreesWithNode_122 (0s).
- runtime-typeof: TestNativeAgreesWithNode_091/internal/oracle/testdata/scout_tsc_nullish_content.a (2.55s), TestNativeAgreesWithNode_091 (0s).
- shared-view: TestNullableObjectStorageArray (2.52s).
- string-check: TestNativeAgreesWithNode_073/internal/oracle/testdata/scout_string_number_stale_field.a (2.87s), TestNativeAgreesWithNode_073 (0s).
- string-tag: TestNativeAgreesWithNode_029/internal/oracle/testdata/scout_string_number_fields_locals.a (2.56s), TestNativeAgreesWithNode_029 (0s).

The recorded IR field-check removal is held by TestNullableStringFieldCheckMutant: the checked baseline panics with exit 70; the unchecked JavaScript mutant prints undefined and exits 0. Source mutant errors are in mutants/results.json and the logs; Go compilation, clang diagnostics, panics and timeouts are not counted as semantic catchers.

Counts additions:

| internal/oracle/testdata/scout_boxed_scalar_fields.a | 62 | 62 | 114 | 169 | 6 | 0 |
| internal/oracle/testdata/scout_boxed_scalar_stale_field.a | 2 | 1 | 3 | 3 | 2 | 0 |
| internal/oracle/testdata/scout_class_boolean_fields.a | 22 | 22 | 9 | 31 | 4 | 0 |
| internal/oracle/testdata/scout_hir_optional_boolean.a | 1 | 1 | 0 | 1 | 1 | 0 |
| internal/oracle/testdata/scout_nullable_object_null_field.a | 2 | 2 | 11 | 12 | 2 | 0 |
| internal/oracle/testdata/scout_nullable_object_stale_field.a | 2 | 1 | 6 | 5 | 2 | 0 |
| internal/oracle/testdata/scout_nullable_object_stale_local.a | 1 | 1 | 6 | 5 | 1 | 0 |
| internal/oracle/testdata/scout_nullable_objects.a | 31 | 31 | 139 | 162 | 5 | 0 |
| internal/oracle/testdata/scout_nullable_string_stale_field.a | 1 | 0 | 4 | 3 | 1 | 0 |
| internal/oracle/testdata/scout_nullable_strings.a | 34 | 34 | 130 | 156 | 5 | 0 |
| internal/oracle/testdata/scout_selector_optional_boolean.a | 2 | 2 | 1 | 4 | 2 | 0 |
| internal/oracle/testdata/scout_string_number_fields_locals.a | 157 | 157 | 137 | 295 | 5 | 0 |
| internal/oracle/testdata/scout_string_number_stale_field.a | 2 | 0 | 3 | 3 | 2 | 0 |
| internal/oracle/testdata/scout_tsc_nullish_content.a | 0 | 0 | 6 | 6 | 0 | 0 |
| internal/oracle/testdata/scout_tsc_source_content.a | 0 | 0 | 2 | 2 | 0 | 0 |
| internal/oracle/testdata/scout_values_optional_boolean.a | 2 | 2 | 2 | 5 | 2 | 0 |

Existing changed row:
{
  "internal/oracle/testdata/taste_stage3_representations.a": {
    "before": "| internal/oracle/testdata/taste_stage3_representations.a | 24 | 24 | 11 | 34 | 10 | 0 |",
    "after": "| internal/oracle/testdata/taste_stage3_representations.a | 24 | 24 | 13 | 36 | 10 | 0 |"
  }
}
