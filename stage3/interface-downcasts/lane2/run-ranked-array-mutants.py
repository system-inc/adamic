#!/usr/bin/env python3
"""Execution guards for the next five ranked array contracts."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path('/tmp/adamic-lane2-ranked-array-mutants');logs.mkdir(exist_ok=True)
cases=[
 ('native-array-kind','internal/native/runtime/object.c',
  'unsigned char actual = adamic_object_field_types(owner)[cache->index];',
  'unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 5) return *slot;', 'ranked-union-wrong-array'),
 ('javascript-array-kind','internal/javascript/readiness.go',
  'type === 5 ? Array.isArray(value)','type === 5 ? true','ranked-union-wrong-array'),
 ('javascript-descendant-kind','internal/javascript/readiness.go',
  '(type === 1 || type === 7) ? typeof value === "number"','(type === 1 || type === 7) ? true','ranked-intersection-wrong-field'),
 ('native-eager-elements','internal/native/runtime/object.c',
  'if (reference != NULL && reference->kind == kind) { return *slot; }',
  'if (reference != NULL && reference->kind == kind) { if (wanted == 5) { const adamic_array *array = slot->reference; for (size_t index = 0; index < array->length; index++) { adamic_slot_cache cache = {NULL, 0}; (void)adamic_object_view(array->elements[index].reference, "id", &cache, 1, "number", "<eager array scan>.id"); } } return *slot; }', 'ranked-union-lazy'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRankedArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
