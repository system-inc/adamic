#!/usr/bin/env python3
"""Prove actual candidate source guards independently, restoring each mutation."""
from pathlib import Path
import subprocess,os
root=Path(__file__).resolve().parents[3]
logs=root/'stage3/interface-downcasts/untagged/logs'
c=root/'internal/native/runtime/view_unions_untagged.c'
js=root/'internal/javascript/view_unions_untagged.go'
test=root/'internal/oracle/checked_views_untagged_unions_test.go'
ir=root/'internal/ir/view_unions_untagged.go'
changes=[
 ('candidate-skip-native',c,'if (adamic_view_untagged_plain_matches(contracts, count, root->members[i], &value))','if (true)','8/wrong'),
 ('candidate-skip-javascript',js,'if(matches(value,member)) return member;','if(true) return member;','8/wrong'),
 ('candidate-wrong-type-native',c,'wanted == adamic_view_union_unknown || value->kind != wanted','wanted == adamic_view_union_unknown','8/wrong'),
 ('candidate-wrong-type-javascript',js,'if(typeof value!==wanted) return false;','/* mutant: trust kind storage */','8/wrong'),
 ('candidate-wrong-literal-native',c,'if (tag_matches(value, &contract->allowed[i]))','if (true)','13/wrong'),
 ('candidate-wrong-literal-javascript',js,"return !contract.Allowed?.length || contract.Allowed.some(literal=>value===(literal.Of===1?literal.Number:literal.Of===2?literal.Boolean:literal.String));",'return true;','13/wrong'),
 ('candidate-drop-transitive',test,'// untagged candidate mutation anchor','if variant=="nested" {dropUntaggedNestedSourceReads(program)}','8/nested'),
 ('candidate-ignore-overlap',ir,' || seen[literal]','','ir'),
 ('candidate-drop-index-membership',test,'// untagged candidate mutation anchor','if variant=="wrong" {dropUntaggedIndexedSourceReads(program)}','133/wrong'),
]
original={p:p.read_bytes() for _,p,_,_,_ in changes}
try:
 for name,path,before,after,case in changes:
  source=original[path].decode();assert source.count(before)==1,(name,'anchor')
  path.write_text(source.replace(before,after))
  package='./internal/ir' if case=='ir' else './internal/oracle'
  pattern='^TestViewUnionDiscriminantOverlaps$' if case=='ir' else '^TestCheckedViewUntaggedCandidatePairs$/'+case
  with (logs/(name+'.log')).open('wb') as output:
   result=subprocess.run(['go','test',package,'-run',pattern,'-count=1','-v'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL: Test' in observed,(name,observed)
  assert '[build failed]' not in observed and 'clang failed' not in observed,(name,observed)
  print(name+': caught by semantic source/IR pin',flush=True)
  path.write_bytes(original[path])
finally:
 for path,data in original.items():path.write_bytes(data)
