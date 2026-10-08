#!/usr/bin/env python3
"""Execution guards for the eighth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked8-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-required-presence','internal/native/runtime/object.c','\tif (slot == NULL || !adamic_object_initialized(object)[cache->index]) {','\tif (slot != NULL && !adamic_object_initialized(object)[cache->index]) {','ranked8-generic-missing'),
 ('javascript-required-presence','internal/javascript/readiness.go','if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic(fieldView','if (adamicFieldReadiness.get(object)?.has(name)) panic(fieldView','ranked8-generic-missing'),
 ('native-record-write-pairs','internal/native/runtime/view_array_writes.c','if (!compatible) array_reference_failure','if (false) array_reference_failure','ranked8-generic-push-wrong'),
 ('javascript-record-write-pairs','internal/javascript/view_array_writes.go','if (!adamicArrayReferencePairs.some(pair => pair[0] === source && pair[1] === target)) adamicArrayReferenceFailure','if (false) adamicArrayReferenceFailure','ranked8-generic-push-wrong'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked8ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
