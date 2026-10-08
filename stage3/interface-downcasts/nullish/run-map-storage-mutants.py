#!/usr/bin/env python3
"""Run independent storage mutations, restoring every source file on exit."""
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
log_root = Path('/tmp/map-storage-mutants')
log_root.mkdir(exist_ok=True)
runtime = 'internal/native/runtime/map_view_storage.c'
cases = [
 ('nullable-slot-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalSlotMutants$/^producer$'),
 ('nullable-slot-js-producer', 'internal/javascript/view_maps_nominal.go', '"((entry) => {if (!(" + strings.Join', '"((entry) => {if (false && !(" + strings.Join', '^TestCheckedViewNullableNominalSlotMutants$/^producer$'),
 ('nullable-slot-native-write', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalSlotMutants$/^write$'),
 ('nullable-slot-js-write', 'internal/javascript/view_maps_nominal.go', '"((entry) => {if (!(" + strings.Join', '"((entry) => {if (false && !(" + strings.Join', '^TestCheckedViewNullableNominalSlotMutants$/^write$'),
 ('nullable-slot-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalSlotMutants$/^read$'),
 ('nullable-slot-js-read', 'internal/javascript/view_nominal_reads.go', '"((value) => (" + strings.Join', '"((value) => (true || " + strings.Join', '^TestCheckedViewNullableNominalSlotMutants$/^read$'),
 ('nullable-slot-null-containment', 'internal/ir/view_writes.go', 'present.Null = present.Null || contract.Null', 'present.Null = false', '^TestCheckedViewNullableNominalSlotMutants$/^narrow-slot$'),
 ('nullable-slot-unboxed-boundary', 'internal/lower/class.go', 'nominalUnion && rawValue.Type() == ir.Object && l.includesNull', 'false && nominalUnion && rawValue.Type() == ir.Object && l.includesNull', '^TestCheckedViewNullableNominalSlotBoundaryRefusals$'),

 ('optional-mutable-late-helper', 'internal/lower/class.go', 'write.WriteContract = writeContract', 'write.WriteContract = 0', '^TestCheckedViewOptionalMutableNominalMutants$/^narrow-slot$'),
 ('optional-mutable-slot', 'internal/ir/view_writes.go', 'present.Undefined = present.Undefined || contract.Undefined', 'present.Undefined = false', '^TestCheckedViewOptionalMutableNominalMutants$/^narrow-slot$'),
 ('optional-mutable-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewOptionalMutableNominalMutants$/^producer$'),
 ('optional-mutable-js-producer', 'internal/javascript/view_maps_nominal.go', '"((entry) => {if (!(" + strings.Join', '"((entry) => {if (false && !(" + strings.Join', '^TestCheckedViewOptionalMutableNominalMutants$/^producer$'),
 ('optional-mutable-native-write', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewOptionalMutableNominalMutants$/^write$'),
 ('optional-mutable-js-write', 'internal/javascript/view_maps_nominal.go', '"((entry) => {if (!(" + strings.Join', '"((entry) => {if (false && !(" + strings.Join', '^TestCheckedViewOptionalMutableNominalMutants$/^write$'),
 ('optional-mutable-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewOptionalMutableNominalMutants$/^read$'),
 ('optional-mutable-js-read', 'internal/javascript/view_nominal_reads.go', '"((value) => (" + strings.Join', '"((value) => (true || " + strings.Join', '^TestCheckedViewOptionalMutableNominalMutants$/^read$'),
 ('nullable-elements-storage', 'internal/ir/view_maps_nominal.go', 'return from.NominalClass != 0 && to.NominalClass != 0', 'return false && from.NominalClass != 0 && to.NominalClass != 0', '^TestCheckedViewMapCertificates$/entry-nominal-elements-.*-widen$'),
 ('nullable-elements-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalArrayMutants$/.*-read$'),
 ('nullable-elements-js-read', 'internal/javascript/view_nominal_reads.go', '"((value) => (" + strings.Join', '"((value) => (true || " + strings.Join', '^TestCheckedViewNullableNominalArrayMutants$/.*-read$'),
 ('nullable-elements-native-write', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalArrayMutants$/.*-(write|widen)$'),
 ('nullable-elements-js-write', 'internal/javascript/view_array_writes.go', 'if (!(adamicInstanceOf(value,adamicArrayNominalClasses[target])', 'if (false && !(adamicInstanceOf(value,adamicArrayNominalClasses[target])', '^TestCheckedViewNullableNominalArrayMutants$/.*-(write|widen)$'),
 ('nullable-elements-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalArrayMutants$/.*-producer$'),
 ('nullable-elements-js-producer', 'internal/javascript/view_maps_nominal.go', '"((entry) => {if (!(" + strings.Join', '"((entry) => {if (false && !(" + strings.Join', '^TestCheckedViewNullableNominalArrayMutants$/.*-producer$'),
 ('json-array-native-representation', 'internal/native/json_array_reads.go', 'wanted := read.Element', 'wanted := read.Element\n if wanted == ir.Map || wanted == ir.Closure { wanted = ir.Object }', '^TestCheckedViewJSONArrayDomainMutants$/^(maps|functions)$'),
 ('json-array-generic-concrete', 'internal/lower/library_json_stringify.go', 'element = l.concrete(element)', '// Counterfactual: keep the uninstantiated element type.', '^TestCheckedViewArrayJSONStorage$/^generic$'),
 ('json-array-native-snapshot-owner', 'internal/native/runtime/json_stringify.c', 'adamic_release(owner); /* JSON element snapshots. */', '/* Counterfactual: retain JSON element snapshots. */', '^TestCheckedViewArrayJSONStorage$/^(finite|nested|arguments)$'),
 ('json-array-native-domain', 'internal/native/json_array_reads.go', 'if (!(%s)) adamic_nullish_failure', 'if (false && !(%s)) adamic_nullish_failure', '^TestCheckedViewJSONArrayDomainMutants$'),
 ('json-array-javascript-domain', 'internal/javascript/library_json_stringify.go', 'check = e.viewArrayChecker(schema.ArrayRead)', 'check = "(value) => value"', '^TestCheckedViewJSONArrayDomainMutants$'),
 ('optional-nominal-javascript-object-kind', 'internal/javascript/view_maps_nominal.go', " && !Array.isArray(entry) && !(entry instanceof Map) && !(entry instanceof Set)", "", '^TestCheckedViewOptionalNominalMutants$/^container$'),
 ('optional-nominal-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewOptionalNominalMutants$/^producer$'),
 ('optional-nominal-javascript-producer', 'internal/javascript/view_maps_nominal.go', '\"((entry) => {if (!(\" + strings.Join', '\"((entry) => {if (false && !(\" + strings.Join', '^TestCheckedViewOptionalNominalMutants$/^producer$'),
 ('optional-nominal-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewOptionalNominalMutants$/^read$'),
 ('optional-nominal-javascript-read', 'internal/javascript/view_nominal_reads.go', '\"((value) => (\" + strings.Join', '\"((value) => (true || \" + strings.Join', '^TestCheckedViewOptionalNominalMutants$/^read$'),
 ('aggregate-nominal-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNullableNominalAggregateMutants$'),
 ('aggregate-nominal-javascript-producer', 'internal/javascript/view_maps_nominal.go', '\"((entry) => {if (!(\" + strings.Join', '\"((entry) => {if (false && !(\" + strings.Join', '^TestCheckedViewNullableNominalAggregateMutants$'),
 ('mutable-nominal-native-write', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewMutableNominalWriteMutant$'),
 ('mutable-nominal-javascript-write', 'internal/javascript/view_maps_nominal.go', '\"((entry) => {if (!(\" + strings.Join', '\"((entry) => {if (false && !(\" + strings.Join', '^TestCheckedViewMutableNominalWriteMutant$'),
 ('mutable-nominal-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewMutableNominalFieldMutant$'),
 ('mutable-nominal-javascript-read', 'internal/javascript/view_nominal_reads.go', '\"((value) => (\" + strings.Join', '\"((value) => (true || \" + strings.Join', '^TestCheckedViewMutableNominalFieldMutant$'),
 ('array-nominal-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNominalArrayProducerMutant$'),
 ('array-nominal-javascript-producer', 'internal/javascript/view_maps_nominal.go', '\"((entry) => {if (!(\" + strings.Join', '\"((entry) => {if (false && !(\" + strings.Join', '^TestCheckedViewNominalArrayProducerMutant$'),
 ('array-nominal-native-write', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNominalArrayWriteMutant$'),
 ('array-nominal-javascript-write', 'internal/javascript/view_array_writes.go', 'if (!(adamicInstanceOf(value,', 'if (false && !(adamicInstanceOf(value,', '^TestCheckedViewNominalArrayWriteMutant$'),
 ('array-nominal-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewNominalArrayReadMutant$'),
 ('array-nominal-javascript-read', 'internal/javascript/view_nominal_reads.go', '\"((value) => (\" + strings.Join', '\"((value) => (true || \" + strings.Join', '^TestCheckedViewNominalArrayReadMutant$'),
 ('nested-nominal-native-producer', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewMapNestedNominalMutants$/^(constructor|set|copy)$'),
 ('nested-nominal-javascript-producer', 'internal/javascript/view_maps_nominal.go', '\"((entry) => {if (!(\" + strings.Join', '\"((entry) => {if (false && !(\" + strings.Join', '^TestCheckedViewMapNestedNominalMutants$/^(constructor|set|copy)$'),
 ('nested-nominal-native-read', 'internal/native/view_nominal_reads.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewMapNestedNominalMutants$/^(receiver|generic|field)$'),
 ('nested-nominal-javascript-read', 'internal/javascript/view_nominal_reads.go', '\"((value) => (\" + strings.Join', '\"((value) => (true || \" + strings.Join', '^TestCheckedViewMapNestedNominalMutants$/^(receiver|generic|field)$'),
 ('boxed-key-hash', 'internal/native/runtime/map.c', 'if (key->kind == adamic_kind_number) return adamic_map_number_hash(((const adamic_number_box *)key)->number);', 'if (key->kind == adamic_kind_number) return (uint64_t)(uintptr_t)key >> 4;', 'entry-convert-key-boxed'),
 ('boxed-key-zero', 'internal/native/runtime/map.c', 'if (number == 0 && signbit(number))', 'if (false && number == 0 && signbit(number))', 'entry-convert-key-boxed-mixed'),
 ('boxed-key-undefined-query', runtime, 'if (map->key_type == 7) return adamic_map_get(map,(adamic_value){.number=adamic_maybe_number_pack((adamic_maybe_number){false,0})});', 'if (map->key_type == 7) return NULL;', 'entry-convert-key-boxed-optional-number'),
 ('boxed-key-equality', 'internal/native/runtime/map.c', 'return a == b || (isnan(a) && isnan(b));', 'return a == b && left == right;', 'entry-convert-key-boxed'),
 ('boxed-key-nan', 'internal/native/runtime/map.c', 'return a == b || (isnan(a) && isnan(b));', 'return a == b;', 'entry-convert-key-boxed'),
 ('boxed-key-payload', runtime, 'adamic_box_number(value.number)', 'adamic_box_number(0)', 'entry-convert-key-boxed'),
 ('boxed-key-ownership', runtime, 'if (map_read_reference(wanted)) { value.reference = adamic_retain(value.reference); }', '(void)wanted;', 'entry-convert-key-boxed-string'),
 ('packed-array-presence', 'internal/native/runtime/view_arrays.c', 'if (!boolean.present)', 'if (false && !boolean.present)', 'entry-convert-nested-boolean-optional'),
 ('packed-array-pack', 'internal/native/runtime/view_arrays.c', '(adamic_maybe_boolean){true, snapshot->boolean}', '(adamic_maybe_boolean){false, snapshot->boolean}', 'entry-convert-nested-boolean'),
 ('packed-array-fallback', 'internal/native/view_arrays.go', 'fallback = "(adamic_value){.maybe_boolean = 2}"', 'fallback = "(adamic_value){.maybe_boolean = 0}"', 'entry-convert-nested-boolean-optional'),
 ('packed-array-native-literal', 'internal/native/view_arrays.go', 'if (%s != NULL && !(%s)) adamic_view_literal_failure', 'if (false && %s != NULL && !(%s)) adamic_view_literal_failure', '^TestCheckedViewMapNestedUnionLiteralMutant$/^entry-convert-nested-boolean-finite$'),
 ('packed-array-javascript-literal', 'internal/javascript/view_arrays.go', 'if (allowed.length && !allowed.includes(value))', 'if (false && allowed.length && !allowed.includes(value))', '^TestCheckedViewMapNestedUnionLiteralMutant$/^entry-convert-nested-boolean-finite$'),
 ('boxed-native-member', 'internal/native/view_arrays_union.go', 'if (!(%s)) adamic_nullish_failure', 'if (false && !(%s)) adamic_nullish_failure', '^TestCheckedViewMapNestedUnionLiteralMutant$'),
 ('boxed-javascript-member', 'internal/javascript/view_arrays_union.go', 'if (!(%s)) panic', 'if (false && !(%s)) panic', '^TestCheckedViewMapNestedUnionLiteralMutant$'),
 ('boxed-number-payload', 'internal/native/runtime/view_arrays.c', 'adamic_box_number(snapshot->number)', 'adamic_box_number(0)', 'entry-convert-nested-union'),
 ('boxed-owner', 'internal/native/runtime/view_arrays.c', 'adamic_array_push(owner, *snapshot);', '(void)owner;', 'entry-convert-nested-union'),
 ('nominal-native', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewMapNominalProducerMutants$'),
 ('nominal-javascript', 'internal/javascript/view_maps_nominal.go', '"((entry) => {if (!(" + strings.Join', '"((entry) => {if (false && !(" + strings.Join', '^TestCheckedViewMapNominalProducerMutants$'),
 ('key-presence', runtime, 'if (!query.present) { return NULL; }', 'if (false && !query.present) { return NULL; }', 'entry-convert-key'),
 ('key-read-presence', runtime, '(adamic_maybe_number){true, value.number}', '(adamic_maybe_number){false, value.number}', 'entry-convert-key'),
 ('nested-read-guard', 'internal/native/runtime/view_arrays.c', 'if (actual != wanted &&', 'if (false && actual != wanted &&', '^TestCheckedViewMapNestedPayloadMutant$'),
 ('optional-reference-admission', 'internal/ir/view_maps.go', 'return allowsNullish(to, ViewUndefined) && accepts(payload, to)', 'return accepts(payload, to)', 'entry-convert-optional-object-schema'),
 ('number-box', runtime, 'adamic_box_number(value.number)', 'NULL', 'entry-convert-number-null'),
 ('boolean-box', runtime, 'value.boolean ? &adamic_box_true : &adamic_box_false', 'value.boolean ? &adamic_box_false : &adamic_box_true', 'entry-convert-boolean-null'),
 ('number-presence', runtime, '(adamic_maybe_number){true, value.number}', '(adamic_maybe_number){false, value.number}', 'entry-convert-number-undefined'),
 ('boolean-presence', runtime, '(adamic_maybe_boolean){true, value.boolean}', '(adamic_maybe_boolean){false, value.boolean}', 'entry-convert-boolean-undefined'),
 ('number-undefined', runtime, 'present.present ? adamic_box_number(present.number) : NULL', 'adamic_box_number(present.present ? present.number : 0)', 'entry-convert-maybe-number'),
 ('boolean-undefined', runtime, '!present.present ? NULL : present.boolean ?', 'present.boolean ?', 'entry-convert-maybe-boolean'),
 ('reference-ownership', runtime, 'if (map_read_reference(wanted)) { value.reference = adamic_retain(value.reference); }', '(void)wanted;', 'entry-convert-owned-get'),
 ('packed-boolean-field', 'internal/native/runtime/view_nullish.c', 'kind=value.present?2:13;boolean=value.boolean;', 'kind=2;boolean=value.boolean;', 'entry-convert-boolean-iterator'),
 ('native-undefined-proof', 'internal/native/union.go', 'if (%s != NULL) adamic_panic', 'if (false && %s != NULL) adamic_panic', None),
 ('javascript-undefined-proof', 'internal/javascript/javascript.go', '"((value) => value === undefined ? undefined : panic(" + quote("undefined storage conversion failed: "+expression.UndefinedWhere+" expected undefined, found present reference") + "))(" + e.value(expression.Value) + ")"', '"((value) => undefined)(" + e.value(expression.Value) + ")"', None),
]
if len(sys.argv) > 1:
 cases = [case for case in cases if case[0] in sys.argv[1:]]
results = []
for name, relative, before, after, fixture in cases:
 path = root / relative
 original = path.read_text()
 if original.count(before) != 1:
  raise RuntimeError(f'{name}: mutation anchor is not unique')
 pattern = '^TestCheckedViewMapUndefinedStorageMutant$' if fixture is None else '^TestCheckedViewMapCertificates$/^' + fixture + '$'
 if fixture is not None and fixture.startswith('^'):
  pattern = fixture
 log_path = log_root / (name + '.log')
 try:
  path.write_text(original.replace(before, after, 1))
  environment = os.environ.copy()
  environment['ADAMIC_GATE_UNCACHED'] = '1'
  with log_path.open('w') as log:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', pattern, '-v', '-count=1', '-timeout', '5m'], cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT)
  output = log_path.read_text()
  caught = result.returncode != 0 and '--- FAIL:' in output and 'clang failed' not in output and '[build failed]' not in output
  results.append({'name': name, 'test': pattern, 'exit': result.returncode, 'caught_by_behavior': caught, 'log': str(log_path)})
  print(name, result.returncode, caught, flush=True)
 finally:
  path.write_text(original)
 (log_root / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
 if not caught:
  raise RuntimeError(f'{name}: not a behavioral kill; inspect {log_path}')
