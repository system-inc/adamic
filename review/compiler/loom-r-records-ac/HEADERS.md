Headers added

35 added headers: 9 refusal reasons and 26 checker codes. Seven existing checker headers remain unchanged. Every program body is byte-identical to 611cbc94.

| Path | First line |
|---|---|
| internal/oracle/testdata/records_refused/records_compare_properties_left.a | `// a-check: refused record read of toString: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map` |
| internal/oracle/testdata/records_refused/records_compare_properties_right.a | `// a-check: refused record read of constructor: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map` |
| internal/oracle/testdata/records_refused/records_environment_boundary.a | `// a-check: refused record read of toString: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map` |
| internal/oracle/testdata/records_refused/records_named_invalidated.a | `// a-check: refused a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven; snapshot the value before the call, or read and recheck its type after the call` |
| internal/oracle/testdata/records_refused/records_narrowed_number.a | `// a-check: refused a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven; snapshot the value before the call, or read and recheck its type after the call` |
| internal/oracle/testdata/records_refused/records_prototype_in.a | `// a-check: refused record membership for constructor: the record's prototype chain isn't modeled; use Object.hasOwn(record, key) for own membership, or a Map` |
| internal/oracle/testdata/records_refused/records_prototype_read.a | `// a-check: refused record read of toString: the record's prototype chain isn't modeled; check Object.hasOwn(record, key) immediately before reading its own value, or use a Map` |
| internal/oracle/testdata/records_refused/records_prototype_set.a | `// a-check: refused record assignment to __proto__: its inherited setter isn't modeled; define an explicit computed own __proto__ data entry before assigning, or use a Map` |
| review/compiler/loom-records-3/stage3-mutant/records/04_own_values.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/08_option_boolean.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/09_option_undefined.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/10_option_wrong_type.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/11_option_string.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/12_nested_entries.a | `// a-check: refused a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)` |
| review/compiler/loom-records-3/stage3-mutant/records/14_group_by.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/15_inherited_read.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/17_compare_missing_object.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/18_compare_missing_scalar.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/19_compare_empty_objects.a | `// a-check: type error TS2345` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/05_paths-own_constructor/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/06_paths-own_toString/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/07_paths-own_hasOwnProperty/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/08_paths-own___proto__/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/09_exports-missing_constructor/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/10_exports-missing_toString/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/11_exports-missing_hasOwnProperty/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/12_exports-missing___proto__/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/13_exports-own_constructor/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/14_exports-own_toString/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/15_exports-own_hasOwnProperty/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/16_exports-own___proto__/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/17_typesVersions-own_constructor/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/18_typesVersions-own_toString/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/19_typesVersions-own_hasOwnProperty/main.a | `// a-check: type error TS2307` |
| review/compiler/loom-records-3/stage3-mutant/records/input-fixtures/20_typesVersions-own___proto__/main.a | `// a-check: type error TS2307` |
