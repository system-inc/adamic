#!/usr/bin/env python3
"""Whole-array membership and nested-read mutations, always restored."""
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
c=root/'internal/native/runtime/view_unions_untagged.c';js=root/'internal/javascript/view_unions_untagged.go';test=root/'internal/oracle/checked_views_untagged_unions_test.go'
changes=[
 ('array-skip-native',c,'if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value))','if (true)','mixed'),
 ('array-skip-js',js,'if(matches(value,member)) return member;','if(true) return member;','mixed'),
 ('array-wrong-member-native',c,'if (!plain_matches_depth(contracts, count, contract->element, &item, depth + 1))','if (false)','mixed'),
 ('array-wrong-member-js',js,'if(!matches(field===undefined?undefined:field.value,contract.Element,depth+1)) return false;','/* mutant: certify any array member */','mixed'),
 ('array-drop-transitive-native-js',test,'// array transitive mutation anchor','dropUntaggedNestedSourceReads(program)','nested'),
]
original={p:p.read_bytes() for _,p,_,_,_ in changes}
try:
 for name,path,before,after,case in changes:
  text=original[path].decode();assert text.count(before)==1,name
  path.write_text(text.replace(before,after))
  log=root/'stage3/interface-downcasts/untagged/logs'/(name+'.log')
  with log.open('wb') as output:
   run=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewUntaggedArrayUnion$/'+case+'$','-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  assert run.returncode!=0 and '--- FAIL: Test' in observed and '[build failed]' not in observed and 'clang failed' not in observed,(name,observed)
  print(name+': semantic refusal pin caught mutation',flush=True)
  path.write_bytes(original[path])
finally:
 for path,data in original.items():path.write_bytes(data)
