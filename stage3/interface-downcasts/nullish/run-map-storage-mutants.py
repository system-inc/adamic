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
