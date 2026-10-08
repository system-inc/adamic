#!/usr/bin/env python3
"""Execution guards for the thirteenth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked13-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-array-element-as-object','internal/native/runtime/view_arrays.c','reference->kind == adamic_kind_object ? 4 : reference->kind == adamic_kind_array ? 5','reference->kind == adamic_kind_object ? 4 : reference->kind == adamic_kind_array ? 4','ranked13-anonymous-alias-arguments-wrong-array'),
 ('javascript-array-element-as-object','internal/javascript/view_arrays.go','type === 4 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) : type === 5 ? Array.isArray(value) : type === 8 ? adamicTypeOf','type === 4 ? value !== null && typeof value === "object" && !(value instanceof Map) : type === 5 ? Array.isArray(value) : type === 8 ? adamicTypeOf','ranked13-anonymous-alias-arguments-wrong-array'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked13ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
