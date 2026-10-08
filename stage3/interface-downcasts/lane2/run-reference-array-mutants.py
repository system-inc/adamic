#!/usr/bin/env python3
"""Execution guards for the certified reference array writes."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path('/tmp/adamic-lane2-reference-array-mutants');logs.mkdir(exist_ok=True)
cases=[
 ('original-slot-alias','internal/lower/view_array_writes.go','write.WriteContract = write.ArraySlotWriteContract','write.WriteContract = 0','reference-array-source-alias'),
 ('native-original-contract','internal/native/runtime/view_array_writes.c','if (!compatible) array_reference_failure(names[target],names[source]);','if (false) array_reference_failure(names[target],names[source]); (void)compatible;','reference-array-missing-field'),
 ('javascript-original-contract','internal/javascript/view_array_writes.go','if (!adamicArrayReferencePairs.some(pair => pair[0] === source && pair[1] === target))','if (false)','reference-array-narrow-field'),
 ('native-incoming-readiness','internal/native/view_array_writes.go','e.line("adamic_verify_array_record(%s);", value)','e.line("if (false) adamic_verify_array_record(%s);", value)','reference-array-uninitialized'),
 ('javascript-incoming-readiness','internal/javascript/view_array_writes.go','for (const field of adamicArrayReferenceFields[source]) adamicViewField','for (const field of []) adamicViewField','reference-array-uninitialized'),
 ('readonly-incoming','internal/ir/view_writes.go','own.Readonly || !assignable(field.Contract, own.Contract, seen)','!assignable(field.Contract, own.Contract, seen)','reference-array-readonly-incoming'),
 ('mutable-field-direction','internal/ir/view_writes.go','own.Readonly || !assignable(field.Contract, own.Contract, seen)','own.Readonly','reference-array-mutable-incoming'),
 ('native-reference-ownership','internal/native/emit_statements.go','value = e.heldReferenceIn(array, value)','if (statement.Element != ir.Object) { value = e.heldReferenceIn(array, value) }','reference-array-alias'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewArrayReferenceWrites/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
