#!/usr/bin/env python3
"""Prove nested producer, logical signature and transitive guards; always restore."""
import os
from pathlib import Path
import subprocess
import sys
root=Path(__file__).resolve().parents[3]
logs=root/'stage3/interface-downcasts/untagged/logs'
logs.mkdir(exist_ok=True)
mutants=[
 ('nested-callable-skip-native','internal/native/runtime/view_unions_untagged.c','if (contract->kind == 5) {','if (contract->kind == 5) { return true;','24/wrong'),
 ('nested-callable-skip-js','internal/javascript/view_unions_untagged.go','if(contract.Kind===5){','if(contract.Kind===5){ return true;','24/wrong'),
 ('nested-callable-wrong-result-native','internal/native/runtime/view_callables_contract.h','if (expected->result != 255 && recorded->result != expected->result) { return false; }','/* mutant trusts mismatched result */','24/wrong'),
 ('nested-callable-wrong-result-js','internal/javascript/view_callables_contract.go','(expected.result===255 || recorded.result===expected.result)','true','24/wrong'),
 ('nested-callable-drop-transitive-native','internal/native/runtime/view_unions_untagged.c','if (!adamic_view_untagged_plain_slot(NULL, value, field->name, &slot) || !plain_matches_depth(contracts, count, field->contract, &slot, depth + 1, path)) { return false; }','if (strcmp(field->name,"symbol") == 0) { continue; }\n        if (!adamic_view_untagged_plain_slot(NULL, value, field->name, &slot) || !plain_matches_depth(contracts, count, field->contract, &slot, depth + 1, path)) { return false; }','24/nested'),
 ('nested-callable-drop-transitive-js','internal/javascript/view_unions_untagged.go','return fields.every(field=>{const actual=slot(value,field.Name);','return fields.every(field=>{if(field.Name===\'symbol\')return true;const actual=slot(value,field.Name);','24/nested'),
 ('nested-callable-drop-logical-native','internal/native/view_callables_signature.go','e.certifyUntaggedCallableRecorded(property, value, recorded)','// mutant trusts physical signature alone','227/logical'),
 ('nested-callable-drop-logical-js','internal/javascript/view_callables_signature.go','recorded = e.untaggedCallableRecorded(property, recorded)','// mutant trusts physical signature alone','227/logical'),
 ('nested-callable-drop-checker-proof','internal/lower/view_unions_untagged.go','if checker.Checker_isTypeIdenticalTo(l.checker, producer.proven, target) {','if target != nil {','227/logical'),
]
if len(sys.argv)>1:mutants=[m for m in mutants if m[0] in sys.argv[1:]]
original={root/p:(root/p).read_bytes() for _,p,_,_,_ in mutants}
try:
 for name,p,before,after,case in mutants:
  path=root/p
  source=original[path].decode()
  assert before in source,(name,'anchor changed')
  # The C recursive field guard occurs in tag and field loops; mutate the last occurrence.
  at=source.rfind(before)
  path.write_text(source[:at]+after+source[at+len(before):])
  rank,variant=case.split('/')
  with (logs/(name+'.log')).open('wb') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run',f'^TestCheckedViewUntaggedCandidatePairs$/{rank}$/{variant}$','-count=1','-v'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  observed=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL: TestCheckedViewUntaggedCandidatePairs' in observed,(name,observed)
  assert '[build failed]' not in observed and 'clang failed' not in observed,(name,observed)
  assert 'exitCode:0' in observed or 'exit=0 stdout="true' in observed,(name,observed)
  print(name+': semantic pinned refusal caught mutation',flush=True)
  path.write_bytes(original[path])
finally:
 for path,content in original.items():path.write_bytes(content)
