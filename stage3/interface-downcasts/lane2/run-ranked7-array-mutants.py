#!/usr/bin/env python3
"""Execution guards for the seventh ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked7-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-field-literal','internal/native/view_fields.go','\t\t\treturn test\n\t\t}(), cString(property.View)','\t\t\treturn "true"\n\t\t}(), cString(property.View)','ranked7-heritage-wrong-token'),
 ('javascript-field-literal','internal/javascript/javascript.go','e.values(expression.ViewAllowed), expression.Absent','"", expression.Absent','ranked7-heritage-wrong-token'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked7ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
