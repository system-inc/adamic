#!/usr/bin/env python3
"""Execution guards for the sixth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked6-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-undefined-element','internal/native/runtime/view_arrays.c','if (reference == NULL) { if (!undefined_allowed) {','if (reference == NULL) { if (false) {','ranked6-jsx-hole'),
 ('native-undefined-element-name','internal/native/runtime/view_arrays.c','if (reference == NULL) { if (!undefined_allowed) { array_view_failure(expression, expected, "undefined");','if (reference == NULL) { if (!undefined_allowed) { array_view_failure(expression, expected, "nullish");','ranked6-jsx-hole'),
 ('javascript-undefined-element','internal/javascript/view_arrays.go','if (value === undefined) { if (required) panic(','if (value === undefined) { if (false) panic(','ranked6-jsx-hole'),
 ('native-join-elements','internal/native/view_arrays.go','slot := e.emitViewArrayRead(read, array, "(double)"+index)','slot := e.temporary(); e.line("adamic_value *%s = adamic_array_holes_at(%s, (double)%s);", slot, array, index)','ranked6-template-join-wrong'),
 ('javascript-join-elements','internal/javascript/view_arrays.go','array.map(value => check(value)).join(separator)','array.join(separator)','ranked6-template-join-wrong'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked6ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
