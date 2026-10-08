#!/usr/bin/env python3
"""Semantic mutants of integrated lowering and native checked reads."""
import pathlib, subprocess
root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path('/tmp/adamic-view-lane2-integrated-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('array-kind', 'internal/native/runtime/object.c', 'unsigned char actual = adamic_object_field_types(owner)[cache->index];', 'unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 5) { return *slot; }', 'TestCheckedViewArrays/array-length-kind'),
 ('element', 'internal/native/runtime/view_arrays.c', 'if (actual != wanted &&', 'if (false && actual != wanted &&', 'TestCheckedViewArrays/array-second'),
 ('signature', 'internal/lower/view_callables.go', 'if signatureProven {', 'if true {', 'TestCheckedViewOpaqueSignature'),
]
for name, relative, before, after, test in cases:
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1
 try:
  path.write_text(original.decode().replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^'+test+'$', '-count=1', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed and '[build failed]' not in observed, observed
  assert 'clang failed' not in observed, observed
  print(name + ': caught by ' + test + '; ' + str(log), flush=True)
 finally:
  path.write_bytes(original)
