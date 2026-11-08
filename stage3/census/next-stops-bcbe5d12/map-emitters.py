import json
from pathlib import Path
out=Path(__file__).resolve().parent
rows=json.loads((out/'source-excerpts.json').read_text());result=[]
for row in rows:
 reason=row['reason'];where=row['where']
 if reason.startswith('a generic instantiation whose type argument for '):
  e=dict(file='internal/lower/constraint_relation.go',line=66,function='(*lowering).constraintRelation',function_line=61,caller='(*lowering).constraintRelationRefusal',caller_line=51,note='constraintRelationRefusal delegates each type-argument binding to constraintRelation. A non-library binding with argument == nil emits this NotYet; the rule is registered at line 56.')
 elif reason.startswith('a namespace read before runtime initialization'):
  e=dict(file='internal/lower/namespaces.go',line=269,function='(*lowering).namespaceInitialization',function_line=260,caller='lowerWithElementKey',caller_file='internal/lower/lower.go',caller_line=104,note='The checkCall closure follows graph.reach(target). A reached namespace declaration absent from initialized emits line 269 at the reached read. The direct-read alternative emits the same text at line 321. lowerWithElementKey invokes the initialization walk before refusal/body lowering.')
 elif reason.startswith('reading an enum before its runtime initialization'):
  e=dict(file='internal/lower/namespaces.go',line=318,function='(*lowering).namespaceInitialization',function_line=260,caller='lowerWithElementKey',caller_file='internal/lower/lower.go',caller_line=104,note='The visit closure sees a namespace value node whose namespaceRuntimeEnum declaration is not in initialized and emits line 318. The reached-call alternative emits the same text at line 267. lowerWithElementKey invokes the initialization walk before refusal/body lowering.')
 elif reason.startswith('a method read as a value ('):
  e=dict(file='internal/lower/refusals.go',line=285,function='(*lowering).methodValueRefusal',function_line=260,caller='(*lowering).refuse',caller_file='internal/lower/refusals.go',caller_line=105,note='The direct-bind exemption is at line 261. The guard at line 264 selects a property-access read outside a direct call without an allowed method-read proof. The method symbol at line 272 selects the refusal emitted at line 285. The refusal registry registers this rule at ordinal 2900, line 291. The visit closure in refuse (function begins at line 35) invokes rule.check at line 105, returning the first rule error. The outer call is Function.prototype.toString.call(func); the toString property is the receiver of .call rather than the direct call expression, so called(node) does not exempt it.')
 elif reason=="a cast the runtime can't check":
  e=dict(file='internal/lower/cast_proof.go',line=191,function='(*lowering).castProof',function_line=178,caller='(*lowering).castRefusal',caller_line=156,note='castRefusal calls castProof at line 161. castProof constructs the exact refusal at line 191 and returns it when no checked cast proof applies.')
 else:raise RuntimeError('Unmapped new group: '+str(row))
 result.append(dict(where=where,kind=row['kind'],reason=reason,**e))
(out/'new-stop-emitters.json').write_text(json.dumps(result,indent=2)+'\n')
print('mapped_groups='+str(len(result)))
