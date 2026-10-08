#!/usr/bin/env python3
"""Active recursive-pair and semantic-undefined-write mutations, restored."""
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
changes=[
 ('cycle-drop-active-native','internal/native/runtime/view_unions_untagged.c','if (seen->contract == id && seen->reference == value->payload.reference)','if (false)','^TestCheckedViewUntaggedRecursive$/cycle$'),
 ('cycle-drop-active-js','internal/javascript/view_unions_untagged.go','if(seen?.has(id)) return true;','/* mutant: omit active pair */','^TestCheckedViewUntaggedRecursive$/cycle$'),
 ('undefined-drop-reference-write','internal/native/runtime/object.c','(actual == 13 && reference_write && object->shape->references[cache->index])','(false && reference_write)','^TestCheckedViewUntaggedOwnClassData$/good$'),
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
