#!/usr/bin/env python3
"""Execution guards for the ninth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked9-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-search-elements','internal/native/view_array_search.go','slot := e.emitViewArrayRead(search.ViewRead.Index(), array, index)','slot := e.temporary(); e.line("adamic_value *%s = adamic_array_holes_at(%s, %s);", slot, array, index)','ranked9-type-stack-includes-wrong'),
 ('javascript-search-elements','internal/javascript/view_array_search.go','const value = index in array ? check(array[index]) : undefined;','const value = array[index];','ranked9-type-stack-includes-wrong'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked9ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
