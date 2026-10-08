#!/usr/bin/env python3
"""Run executable NodeArray guard mutants; always restore production files."""
import pathlib, subprocess, os, sys
root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path('/tmp/adamic-lane2-node-array-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('native-element-kind', 'internal/native/runtime/view_arrays.c',
  'if (actual != wanted && !(wanted == 7 && actual == 1) && !(wanted == 10 && array->element_kind == 10))',
  'if (false)', 'node-array-wrong-element'),
 ('javascript-element-kind', 'internal/javascript/view_arrays.go',
  'if (!valid) panic("element read failed:', 'if (false) panic("element read failed:', 'node-array-wrong-element'),
 ('native-own-kind', 'internal/native/runtime/object.c',
  '\tunsigned char actual = adamic_object_field_types(owner)[cache->index];',
  '\tunsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 1) return *slot;', 'node-array-wrong-own'),
 ('native-property-copy', 'internal/native/view_array_records.go',
  'adamic_object_copy_checked(%s, %s);", array, properties, cString(record.Where)',
  'adamic_retain(%s);", array, properties', 'node-array-properties-copy'),
 ('javascript-own-kind', 'internal/javascript/readiness.go',
  '(type === 1 || type === 7) ? typeof value === "number"',
  '(type === 1 || type === 7) ? true', 'node-array-wrong-own'),
 ('source-slot-certificate', 'internal/lower/view_array_records.go',
  'l.optionalViewWriteField(name)', '_ = name', 'node-array-write-literal'),
]
for name, relative, before, after, probe in cases:
 if len(sys.argv) > 1 and name not in sys.argv[1:]: continue
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1, name
 try:
  path.write_text(original.decode().replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewNodeArrayRecords/' + probe + '$', '-count=1', '-timeout', '10m'], cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED':'1'}, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed, observed
  assert all(marker not in observed for marker in ['[build failed]', 'clang failed', 'SyntaxError', 'compiler bug:', "can't lower"]), observed
  print(name + ': caught by ' + probe + '; ' + str(log), flush=True)
 finally:
  path.write_bytes(original)
