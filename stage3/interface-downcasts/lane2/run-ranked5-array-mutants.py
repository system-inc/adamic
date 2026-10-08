#!/usr/bin/env python3
"""Execution guards for the fifth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked5-array-mutants';logs.mkdir(exist_ok=True)
cases=[
 ('native-array-kind','internal/native/runtime/object.c','unsigned char actual = adamic_object_field_types(owner)[cache->index];\n\t// Boxed','unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 5) return *slot;\n\t// Boxed','ranked5-signatures-wrong-array'),
 ('javascript-array-kind','internal/javascript/readiness.go','type === 5 ? Array.isArray(value)','type === 5 ? value !== null && typeof value === "object"','ranked5-tuple-wrong-array'),
 ('native-element-kind','internal/native/runtime/view_arrays.c','if (actual != wanted && !(wanted == 7 && actual == 1)','if (false && actual != wanted && !(wanted == 7 && actual == 1)','ranked5-tuple-wrong-element'),
 ('javascript-element-kind','internal/javascript/view_arrays.go','if (!valid) panic("element read failed: "','if (false) panic("element read failed: "','ranked5-jsdoc-wrong-element'),
 ('native-union-element-tag','internal/native/view_unions.go','e.line("if (!(%s)) adamic_view_literal_failure(','e.line("if (false && !(%s)) adamic_view_literal_failure(','ranked5-modifiers-wrong-tag'),
 ('javascript-union-element-tag','internal/javascript/readiness.go','    if (allowed.length && !allowed.includes(value)) panic(','    if (false) panic(','ranked5-modifiers-wrong-tag'),
 ('native-nullable-member-kinds','internal/native/runtime/view_nullish.c','if(kind==0||(kinds&(1u<<kind))==0)failure','if(kind==0)failure','ranked5-heritage-wrong-array'),
 ('javascript-nullable-member-kinds','internal/javascript/readiness.go','(kinds & (1 << kind)) !== 0 &&','true &&','ranked5-modifiers-wrong-array'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked5ArrayContracts/'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:',"can't lower"]),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
