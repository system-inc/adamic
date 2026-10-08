#!/usr/bin/env python3
"""Execution guards for the fourth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path('/workspace/adamic')
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked4-array-mutants';logs.mkdir(exist_ok=True)
cases=[('native-graph-object-size', 'internal/native/graph_regions.go', 'fmt.Sprintf("adamic_object_size(%s->shape->count)", value)', 'fmt.Sprintf("sizeof *%s + %s->shape->count * (sizeof(adamic_value) + 2)", value, value)', 'ranked4-comma'), ('native-write-type-name', 'internal/native/emit_statements.go', ', ir.Object: "object", ir.Array: "array", ir.Map: "Map"', '', 'ranked4-label-assign'), ('native-record-write-pairs', 'internal/native/runtime/view_array_writes.c', 'if (!compatible) array_reference_failure', 'if (false) array_reference_failure', 'ranked4-chain-set-wrong'), ('javascript-record-write-pairs', 'internal/javascript/view_array_writes.go', 'if (!adamicArrayReferencePairs.some(pair => pair[0] === source && pair[1] === target)) adamicArrayReferenceFailure', 'if (false) adamicArrayReferenceFailure', 'ranked4-chain-set-wrong'), ('native-scalar-write-storage', 'internal/native/runtime/view_arrays.c', 'if (array->element_kind != physical) { array_view_failure', 'if (false) { array_view_failure', 'ranked4-files-push-wrong'), ('javascript-scalar-write-storage', 'internal/javascript/view_arrays.go', 'if (actual !== (storage === 1 || storage === 2 || storage === 7 ? storage : 10)) panic', 'if (false) panic', 'ranked4-files-set-wrong'), ('native-undefined-array-write-refusal', 'internal/native/runtime/object.c', 'wanted == 3 || wanted == 4 || wanted == 6 || wanted == 8 || wanted == 10', '(wanted >= 3 && wanted <= 6) || wanted == 8 || wanted == 10', 'ranked4-label-assign')]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked4ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
