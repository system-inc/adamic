#!/usr/bin/env python3
"""Array-only semantic mutants, restoring each file even on a failed run."""
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path('/tmp/adamic-lane2-native-array-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('field-array-kind', 'internal/native/runtime/object.c', 'unsigned char actual = adamic_object_field_types(owner)[cache->index];', 'unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 5) { return *slot; }', 'legacy:array-length-kind'),
 ('allocation-kind', 'internal/native/emit_expressions.go', 'e.line("%s->element_kind = %d;", array, kind)', 'e.line("%s->element_kind = %d;", array, ir.Type(0))', 'native-array-join'),
 ('element-kind', 'internal/native/runtime/view_arrays.c', 'if (actual != wanted &&', 'if (false && actual != wanted &&', 'native-array-copy-bad'),
 ('write-kind', 'internal/native/runtime/view_arrays.c', 'if (array->element_kind != physical)', 'if (false && array->element_kind != physical)', 'native-array-write-bad'),
 ('push-kind', 'internal/native/runtime/view_arrays.c', 'if (array->element_kind != physical)', 'if (false && array->element_kind != physical)', 'native-array-push-bad'),
 ('literal', 'internal/native/view_arrays.go', 'if len(read.ViewAllowed) != 0 {', 'if false && len(read.ViewAllowed) != 0 {', 'native-array-join-literal'),
 ('copy-kind', 'internal/native/runtime/array.c', 'sliced->element_kind = array->element_kind;', 'sliced->element_kind = 0;', 'native-array-copy-bad'),
 ('join-contract', 'internal/lower/view_arrays.go', 'ViewRead: l.viewArrayUse(node, receiver, element, false)', 'ViewRead: ir.ArrayViewRead{}', 'native-array-join-bad'),
 ('sparse-map', 'internal/native/library_array_holes.go', 'e.line("if (%s == NULL) continue;", slot)', 'e.line("if (false && %s == NULL) continue;", slot)', 'native-array-sparse'),
 ('javascript-write-kind', 'internal/javascript/view_arrays.go', 'if (actual !== (storage === 1', 'if (false && actual !== (storage === 1', 'native-array-write-bad'),
 ('javascript-pop-hole', 'internal/javascript/view_arrays.go', 'index in array ? check(array[index]) : undefined', 'check(array[index])', 'native-array-sparse-pop-hole'),
]
for name, relative, before, after, probe in cases:
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1, (name, before)
 try:
  path.write_text(original.decode().replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   test = ('TestCheckedViewArrays/' + probe.removeprefix('legacy:')) if probe.startswith('legacy:') else ('TestCheckedViewNativeArrays/' + probe)
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^'+test+'$', '-count=1', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed and '[build failed]' not in observed, observed
  assert 'clang failed' not in observed and 'SyntaxError' not in observed, observed
  print(name + ': caught by ' + probe + '; ' + str(log), flush=True)
 finally:
  path.write_bytes(original)
