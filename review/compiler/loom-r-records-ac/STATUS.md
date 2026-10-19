Per-file a-check status

Compiler: this branch’s own `cmd/adamic`, invoked as `adamic c <path>` by unchanged `Gate.aCheck` at `fbac28c62493f27a788edc02a18bc8edb68de5da`.

109 files: 64 checked, 3 NotYet, 9 refused, 33 type errors. NotYet is separated here although Gate counts it as checked.

| Path | Status | Diagnostic reason or code | Header action |
|---|---|---|---|
| internal/oracle/testdata/detached_own_objects.a | checked |  | none |
| internal/oracle/testdata/detached_own_parser.a | checked |  | none |
| internal/oracle/testdata/detached_own_records.a | checked |  | none |
| internal/oracle/testdata/detached_own_unready.a | checked |  | none |
| internal/oracle/testdata/json_replacer_duplicates_object.a | checked |  | none |
| internal/oracle/testdata/records_buckets/17_compare_missing_object.a | type error | TS2345 | retained |
| internal/oracle/testdata/records_buckets/18_compare_missing_scalar.a | type error | TS2345 | retained |
| internal/oracle/testdata/records_buckets/19_compare_empty_objects.a | type error | TS2345 | retained |
| internal/oracle/testdata/records_census.a | checked |  | none |
| internal/oracle/testdata/records_compare_missing_scalar.a | checked |  | none |
| internal/oracle/testdata/records_discarded.a | checked |  | none |
| internal/oracle/testdata/records_for_in.a | checked |  | none |
| internal/oracle/testdata/records_guarded_snapshot.a | checked |  | none |
| internal/oracle/testdata/records_json_replacer_duplicates.a | checked |  | none |
| internal/oracle/testdata/records_named_alias_unready.a | checked |  | none |
| internal/oracle/testdata/records_named_intrinsics.a | checked |  | none |
| internal/oracle/testdata/records_named_operations.a | checked |  | none |
| internal/oracle/testdata/records_named_options.a | checked |  | none |
| internal/oracle/testdata/records_named_union.a | checked |  | none |
| internal/oracle/testdata/records_narrowed_reference.a | checked |  | none |
| internal/oracle/testdata/records_operations.a | checked |  | none |
| internal/oracle/testdata/records_optional.a | checked |  | none |
| internal/oracle/testdata/records_ownership.a | checked |  | none |
| internal/oracle/testdata/records_partial_parser.a | checked |  | none |
| internal/oracle/testdata/records_partial_undefined.a | checked |  | none |
| internal/oracle/testdata/records_partial_views.a | checked |  | none |
| internal/oracle/testdata/records_refuse/r24.a | NotYet | String conversion of an object, array, map or function (ToPrimitive is not lowered) yet | none |
| internal/oracle/testdata/records_refused/records_compare_properties_left.a | refused | record read of toString: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map | added |
| internal/oracle/testdata/records_refused/records_compare_properties_right.a | refused | record read of constructor: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map | added |
| internal/oracle/testdata/records_refused/records_environment_boundary.a | refused | record read of toString: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map | added |
| internal/oracle/testdata/records_refused/records_named_invalidated.a | refused | a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven; snapshot the value before the call, or read and recheck its type after the call | added |
| internal/oracle/testdata/records_refused/records_narrowed_number.a | refused | a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven; snapshot the value before the call, or read and recheck its type after the call | added |
| internal/oracle/testdata/records_refused/records_prototype_in.a | refused | record membership for constructor: the record's prototype chain isn't modeled; use Object.hasOwn(record, key) for own membership, or a Map | added |
| internal/oracle/testdata/records_refused/records_prototype_read.a | refused | record read of toString: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map | added |
| internal/oracle/testdata/records_refused/records_prototype_set.a | refused | record assignment to __proto__: its inherited setter isn't modeled; define an explicit computed own __proto__ data entry before assigning, or use a Map | added |
| internal/oracle/testdata/records_scalars.a | checked |  | none |
| internal/oracle/testdata/records_type_only_composition.a | checked |  | none |
| internal/oracle/testdata/records_type_only_scanner.a | checked |  | none |
| internal/oracle/testdata/records_type_only_value.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/01_has_property.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/02_get_property.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/03_own_keys.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/04_own_values.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/05_integer_order.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/06_delete_readd.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/07_optional_view.a | NotYet | a Record<string, string> seen as Named (fixed objects and records have different storage; copy explicitly) yet | none |
| review/compiler/loom-records-3/stage3-mutant/records/08_option_boolean.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/09_option_undefined.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/10_option_wrong_type.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/11_option_string.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/12_nested_entries.a | refused | a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) | added |
| review/compiler/loom-records-3/stage3-mutant/records/13_strict_option.a | NotYet | a record value of type boolean \| undefined yet | none |
| review/compiler/loom-records-3/stage3-mutant/records/14_group_by.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/15_inherited_read.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/16_built_strings.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/17_compare_missing_object.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/18_compare_missing_scalar.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/19_compare_empty_objects.a | type error | TS2345 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/01_paths-missing_constructor/main.a | type error | TS2307 | retained |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/01_paths-missing_constructor/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/02_paths-missing_toString/main.a | type error | TS2307 | retained |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/02_paths-missing_toString/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/03_paths-missing_hasOwnProperty/main.a | type error | TS2307 | retained |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/03_paths-missing_hasOwnProperty/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/04_paths-missing___proto__/main.a | type error | TS2307 | retained |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/04_paths-missing___proto__/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/05_paths-own_constructor/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/05_paths-own_constructor/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/06_paths-own_toString/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/06_paths-own_toString/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/07_paths-own_hasOwnProperty/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/07_paths-own_hasOwnProperty/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/08_paths-own___proto__/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/08_paths-own___proto__/target.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/09_exports-missing_constructor/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/09_exports-missing_constructor/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/10_exports-missing_toString/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/10_exports-missing_toString/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/11_exports-missing_hasOwnProperty/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/11_exports-missing_hasOwnProperty/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/12_exports-missing___proto__/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/12_exports-missing___proto__/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/13_exports-own_constructor/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/13_exports-own_constructor/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/14_exports-own_toString/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/14_exports-own_toString/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/15_exports-own_hasOwnProperty/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/15_exports-own_hasOwnProperty/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/16_exports-own___proto__/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/16_exports-own___proto__/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/17_typesVersions-own_constructor/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/17_typesVersions-own_constructor/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/18_typesVersions-own_toString/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/18_typesVersions-own_toString/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/19_typesVersions-own_hasOwnProperty/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/19_typesVersions-own_hasOwnProperty/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/20_typesVersions-own___proto__/main.a | type error | TS2307 | added |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/20_typesVersions-own___proto__/node_modules/pkg/target.d.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/21_config-unknown_constructor/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/22_config-unknown_toString/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/23_config-unknown_hasOwnProperty/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/24_config-unknown___proto__/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/25_cli-unknown_constructor/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/26_cli-unknown_toString/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/27_cli-unknown_hasOwnProperty/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/28_cli-unknown___proto__/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/29_compare-old/main.a | checked |  | none |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/30_compare-new/main.a | checked |  | none |
| stage3/records-maplike-main/corePublic-MapLike.a | checked |  | none |
