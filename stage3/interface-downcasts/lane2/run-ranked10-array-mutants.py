#!/usr/bin/env python3
"""Execution guards for the tenth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked10-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-nullable-literal','internal/native/view_nullish.go','e.line("if (%s != NULL && %s != &adamic_null && !(%s)) adamic_nullish_failure(','e.line("if (false && %s != NULL && %s != &adamic_null && !(%s)) adamic_nullish_failure(','ranked10-references-wrong-mode'),
 ('javascript-nullable-literal','internal/javascript/readiness.go','(kinds & (1 << kind)) !== 0 && (!allowed.length || allowed.includes(value))','(kinds & (1 << kind)) !== 0','ranked10-references-wrong-mode'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked10ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
