#!/usr/bin/env python3
"""Recursive source mutations, restored even on failure."""
from pathlib import Path
import subprocess,os
root=Path(__file__).resolve().parents[3]
c=root/'internal/native/runtime/view_unions_untagged.c'
js=root/'internal/javascript/view_unions_untagged.go'
changes=[
 ('recursive-skip-native',c,'if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value))','if (true)','wrong'),
 ('recursive-skip-js',js,'if(matches(value,member)) return member;','if(true) return member;','wrong'),
 ('recursive-wrong-native',c,'wanted == adamic_view_union_unknown || value->kind != wanted','wanted == adamic_view_union_unknown','wrong'),
 ('recursive-wrong-js',js,'if(typeof value!==wanted) return false;','/* mutant */','wrong'),
 ('recursive-drop-nested-native',c,'if (contract->kind == 4) {','if (depth > 0 && contract->kind == 4) { return true; }\n    if (contract->kind == 4) {','nested'),
 ('recursive-drop-nested-js',js,'if(contract.Kind===4) return','if(depth>0 && contract.Kind===4) return true;\n  if(contract.Kind===4) return','nested'),
]
original={p:p.read_bytes() for _,p,_,_,_ in changes}
try:
 for name,path,before,after,case in changes:
  text=original[path].decode();assert text.count(before)==1,name
  path.write_text(text.replace(before,after))
  log=root/'stage3/interface-downcasts/untagged/logs'/(name+'.log')
  with log.open('wb') as output:
   run=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewUntaggedRecursive$/'+case,'-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert run.returncode!=0 and '--- FAIL: Test' in observed and '[build failed]' not in observed and 'clang failed' not in observed,(name,observed)
  print(name+': semantic refusal pin caught mutation',flush=True)
  path.write_bytes(original[path])
finally:
 for path,data in original.items():path.write_bytes(data)
