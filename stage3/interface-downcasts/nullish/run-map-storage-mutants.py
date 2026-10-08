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
 ('packed-json-boundary', 'internal/lower/library_json_stringify.go', 'known && (storage == ir.Union || storage == ir.MaybeBoolean)', 'known && false && (storage == ir.Union || storage == ir.MaybeBoolean)', '^TestCheckedViewArrayJSONStorageRefusal$'),
 ('boxed-native-member', 'internal/native/view_arrays_union.go', 'if (!(%s)) adamic_nullish_failure', 'if (false && !(%s)) adamic_nullish_failure', '^TestCheckedViewMapNestedUnionLiteralMutant$'),
 ('boxed-javascript-member', 'internal/javascript/view_arrays_union.go', 'if (!(%s)) panic', 'if (false && !(%s)) panic', '^TestCheckedViewMapNestedUnionLiteralMutant$'),
 ('boxed-number-payload', 'internal/native/runtime/view_arrays.c', 'adamic_box_number(snapshot->number)', 'adamic_box_number(0)', 'entry-convert-nested-union'),
 ('boxed-owner', 'internal/native/runtime/view_arrays.c', 'adamic_array_push(owner, *snapshot);', '(void)owner;', 'entry-convert-nested-union'),
 ('nominal-native', 'internal/native/view_maps_nominal.go', 'if (!(%s)) adamic_panic', 'if (false && !(%s)) adamic_panic', '^TestCheckedViewMapNominalProducerMutants$'),
 ('nominal-javascript', 'internal/javascript/view_maps_nominal.go', '"((entry) => (" + strings.Join', '"((entry) => (true || " + strings.Join', '^TestCheckedViewMapNominalProducerMutants$'),
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
