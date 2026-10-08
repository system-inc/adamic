#!/usr/bin/env python3
"""Execution guards for the ranked readonly array union contracts."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path('/tmp/adamic-lane2-ranked-array-union-mutants');logs.mkdir(exist_ok=True)
cases=[
 ('native-array-kind','internal/native/runtime/object.c','unsigned char actual = adamic_object_field_types(owner)[cache->index];','unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 5) return *slot;','ranked3-pattern-wrong-array'),
 ('javascript-array-kind','internal/javascript/readiness.go','type === 5 ? Array.isArray(value)','type === 5 ? true','ranked3-pattern-wrong-array'),
 ('native-element-kind','internal/native/runtime/view_arrays.c','if (actual != wanted &&','if (false && actual != wanted &&','ranked3-pattern-wrong-element'),
 ('javascript-element-kind','internal/javascript/view_arrays.go','if (!valid) panic("element read failed:','if (false && !valid) panic("element read failed:','ranked3-pattern-wrong-element'),
 ('native-element-tag','internal/native/view_arrays.go','if read.Element == ir.Object && read.ViewContract != 0 {','if false && read.Element == ir.Object && read.ViewContract != 0 {','ranked3-pattern-wrong-tag'),
 ('javascript-element-tag','internal/javascript/view_arrays.go','if read.Element == ir.Object && read.ViewContract != 0 {\n\t\tchecked =','if false && read.Element == ir.Object && read.ViewContract != 0 {\n\t\tchecked =','ranked3-pattern-wrong-tag'),
 ('union-element-merge','internal/lower/view_array_unions.go','return l.checker.GetUnionType(elements)','return elements[0]','ranked3-pattern-wrong-tag'),
 ('native-long-diagnostic','internal/native/runtime/object.c','strlen(expression) + 2 * strlen(type) + strlen(found) + 100','strlen(expression) + strlen(type) + strlen(found) + 100','ranked3-pattern-wrong-array'),
 ('native-eager-elements','internal/native/runtime/object.c','if (reference != NULL && reference->kind == kind) { return *slot; }','if (reference != NULL && reference->kind == kind) { if (wanted == 5) { const adamic_array *array = slot->reference; for (size_t index = 0; index < array->length; index++) { adamic_slot_cache cache = {NULL, 0}; (void)adamic_object_view(array->elements[index].reference, "pos", &cache, 1, "number", "<eager array scan>.pos"); } } return *slot; }','ranked3-pattern-lazy'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==(2 if name == "javascript-element-tag" else 1),name
 try:
  path.write_text(original.decode().replace(before,after,1))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRankedArrayUnionContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
