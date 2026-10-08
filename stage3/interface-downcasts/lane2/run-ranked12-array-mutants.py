#!/usr/bin/env python3
"""Execution guards for the twelfth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked12-array-mutants';logs.mkdir(exist_ok=True)
# Native checks a nullable union twice: the runtime's member-kind mask, then the emitted member
# selection. Either alone still refuses, so the native mutant disables both; the two single
# mutants are recorded as surviving in RANKED12_ARRAYS_REPORT.md.
cases=[
 ('native-union-member-mask-and-selection','internal/native/view_nullish.go',['e.cache(), property.NullishKinds, property.NullAllowed','e.line("if (%s != NULL && %s != &adamic_null) {", value, value)'],['e.cache(), property.NullishKinds|2, property.NullAllowed','e.line("if (false && %s != NULL && %s != &adamic_null) {", value, value)'],'ranked12-jsdoc-comment-wrong'),
 ('javascript-union-member-mask','internal/javascript/view_nullish.go','quote(property.ViewType), property.NullishKinds, propert','quote(property.ViewType), property.NullishKinds|2, propert','ranked12-jsdoc-comment-wrong'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();text=original.decode()
 befores,afters=(before,after) if isinstance(before,list) else ([before],[after])
 for one,other in zip(befores,afters):
  assert text.count(one)==1,name
  text=text.replace(one,other)
 try:
  path.write_text(text)
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked12ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
