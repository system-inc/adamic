#!/usr/bin/env python3
"""Recursive producer and callable member dispatch semantic mutations."""
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
changes=[
 ('flow-drop-undefined-tag','internal/native/emit_objects.go','return 13','return int(value.Type())','^TestCheckedViewUntaggedCandidatePairs$/108$/good$'),
 ('flow-drop-final-refresh','internal/lower/view_lazy.go','l.completeUntaggedRecursiveContracts()','// mutant omits recursive completion','^TestCheckedViewUntaggedCandidatePairs$/92$/good$'),
 ('callable-skip-native','internal/native/view_fields.go','if contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','if false && contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','^TestCheckedViewUntaggedCallableUnion$/wrong$'),
 ('callable-skip-js','internal/javascript/view_callables.go','if contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','if false && contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','^TestCheckedViewUntaggedCallableUnion$/wrong$'),
 ('callable-wrong-member-native','internal/native/view_unions_untagged.go','return "(" + strings.Join(choices, "") + first + ")"','return first','^TestCheckedViewUntaggedCallableUnion$/good-string$'),
 ('callable-wrong-member-js','internal/javascript/view_unions_untagged.go','return choices.find(expected=>','return choices[0] || choices.find(expected=>','^TestCheckedViewUntaggedCallableUnion$/good-string$'),
 ('callable-drop-nested-native','internal/native/view_fields.go','if contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','if false && contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','^TestCheckedViewUntaggedCallableUnion$/nested$'),
 ('callable-drop-nested-js','internal/javascript/view_callables.go','if contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','if false && contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {','^TestCheckedViewUntaggedCallableUnion$/nested$'),
]
original={root/path:(root/path).read_bytes() for _,path,_,_,_ in changes}
try:
 for name,relative,before,after,pattern in changes:
  path=root/relative;text=original[path].decode();assert text.count(before)==1,name
  path.write_text(text.replace(before,after))
  log=root/'stage3/interface-downcasts/untagged/logs'/(name+'.log')
  with log.open('wb') as output:
   run=subprocess.run(['go','test','./internal/oracle','-run',pattern,'-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert run.returncode!=0 and '--- FAIL: Test' in observed and '[build failed]' not in observed and 'clang failed' not in observed,(name,observed)
  print(name+': semantic source pin caught mutation',flush=True)
  path.write_bytes(original[path])
finally:
 for path,data in original.items():path.write_bytes(data)
