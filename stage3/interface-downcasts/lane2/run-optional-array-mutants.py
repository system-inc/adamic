#!/usr/bin/env python3
"""Optional array kind, lazy reads and transitive element-field checks."""
import pathlib, subprocess
root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path('/tmp/adamic-lane2-optional-array-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('native-array-kind', 'internal/native/runtime/object.c',
  'unsigned char actual = adamic_object_field_types(owner)[cache->index];',
  'unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 5) return *slot;',
  'optional-declarations-wrong-array'),
 ('javascript-array-kind', 'internal/javascript/readiness.go',
  'type === 5 ? Array.isArray(value)', 'type === 5 ? true',
  'optional-declarations-wrong-array'),
 ('native-eager-elements', 'internal/native/runtime/object.c',
  'if (reference != NULL && reference->kind == kind) { return *slot; }',
  'if (reference != NULL && reference->kind == kind) { if (wanted == 5) { const adamic_array *array = slot->reference; for (size_t index = 0; index < array->length; index++) { adamic_value snapshot; adamic_value *element = adamic_view_array_at(array, (double)index, false, false, 4, "Declaration", expression, &snapshot); if (element != NULL) { adamic_slot_cache cache = {NULL, 0}; (void)adamic_object_view(element->reference, "name", &cache, 3, "string", "<eager array field scan>.name"); } } } return *slot; }',
  'optional-declarations-lazy'),
 ('javascript-eager-elements', 'internal/javascript/readiness.go',
  'if (allowed.length && !allowed.includes(value))',
  'if (type === 5) for (let index = 0; index < value.length; index++) adamicViewField(adamicViewArrayElement(value[index], expression, 4, "Declaration", [], true), "name", "<eager array field scan>.name", 3, "string");\n    if (allowed.length && !allowed.includes(value))',
  'optional-declarations-lazy'),
 ('javascript-transitive-field', 'internal/javascript/readiness.go',
  'type === 3 ? typeof value === "string"', 'type === 3 ? true',
  'optional-declarations-wrong-field'),
]
for name, relative, before, after, probe in cases:
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1, name
 try:
  path.write_text(original.decode().replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewOptionalDeclarations/' + probe + '$', '-count=1', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed, observed
  assert '[build failed]' not in observed and 'clang failed' not in observed and 'SyntaxError' not in observed and 'compiler bug:' not in observed and "can't lower" not in observed, observed
  print(name + ': caught by ' + probe + '; ' + str(log), flush=True)
 finally:
  path.write_bytes(original)
