#!/usr/bin/env python3
"""Mutate the real source dispatch and lowered nested reads, restoring all files."""
from pathlib import Path
import os,subprocess
root=Path(__file__).resolve().parents[3]
logs=root/'stage3/interface-downcasts/untagged/logs'
c=root/'internal/native/runtime/view_unions_untagged.c'
js=root/'internal/javascript/view_unions_untagged.go'
test=root/'internal/oracle/checked_views_untagged_unions_test.go'
mutants=[
 ('source-skip-native',c,'if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value))','if (true)','binding-name/wrong'),
 ('source-skip-javascript',js,'if(matches(value,member)) return member;','if(true) return member;','binding-name/wrong'),
 ('source-wrong-shape-native',c,'if (tag_matches(value, &contract->allowed[i]))','if (true)','binding-name/wrong'),
 ('source-wrong-shape-javascript',js,"return !contract.Allowed?.length || contract.Allowed.some(literal=>value===(literal.Of===1?literal.Number:literal.Of===2?literal.Boolean:literal.String));",'return true;','binding-name/wrong'),
 ('source-drop-transitive',test,'// untagged source mutation anchor','if variant=="nested" {dropUntaggedNestedSourceReads(program)}','binding-name/nested'),
]
original={p:p.read_bytes() for _,p,_,_,_ in mutants}
try:
 for name,path,before,after,case in mutants:
  source=original[path].decode();assert source.count(before)==1,(name,'anchor')
  path.write_text(source.replace(before,after))
  with (logs/(name+'.log')).open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewUntaggedSourceDispatch$/'+case,'-count=1','-v'],cwd=root,env={**os.environ,'VIEW_UNTAGGED_SOURCE_REQUIRED':'1','ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL: TestCheckedViewUntaggedSourceDispatch' in observed,(name,observed)
  assert '[build failed]' not in observed and 'clang failed' not in observed,(name,observed)
  print(name+': caught by exact source exit/output pin',flush=True)
  path.write_bytes(original[path])
finally:
 for path,data in original.items():path.write_bytes(data)
