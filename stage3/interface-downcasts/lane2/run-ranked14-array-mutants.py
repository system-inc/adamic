#!/usr/bin/env python3
"""Execution guards for the fourteenth ranked array contracts: each mutant must be caught by its named probe."""
import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('TMPDIR','/tmp'))/'adamic-lane2-ranked14-array-mutants';logs.mkdir(exist_ok=True)
# Each case names the test it runs under: the frontier pin is a lowering refusal, not a probe.
cases=[
 ('name-fallback-removed','internal/lower/view_lazy.go','family = unsupportedFields[field]','family = ""','TestCheckedViewRanked14Frontiers/ranked14-text-name-fallback'),
 ('native-boolean-field','internal/native/runtime/object.c','unsigned char actual = adamic_object_field_types(owner)[cache->index];\n\t// Boxed','unsigned char actual = adamic_object_field_types(owner)[cache->index]; if (wanted == 2) return *slot;\n\t// Boxed','TestCheckedViewRanked14ArrayContracts/ranked14-jsdoc-signature-parameters-wrong-bracketed'),
 ('javascript-boolean-field','internal/javascript/readiness.go','(type === 2 || type === 9) ? typeof value === "boolean"','(type === 2 || type === 9) ? true','TestCheckedViewRanked14ArrayContracts/ranked14-jsdoc-signature-parameters-wrong-bracketed'),
]
for name,relative,before,after,probe in cases:
 path=root/relative;original=path.read_bytes();assert original.decode().count(before)==1,name
 try:
  path.write_text(original.decode().replace(before,after))
  log=logs/(name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^'+probe+'$','-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert result.returncode!=0 and '--- FAIL:' in observed,observed
  assert all(marker not in observed for marker in ['[build failed]','clang failed','SyntaxError','compiler bug:']),observed
  print(name+': caught by '+probe+'; '+str(log),flush=True)
 finally:path.write_bytes(original)
