import pathlib, subprocess
root=pathlib.Path('/workspace/adamic')
mutants=[
 ('helper-proof', 'internal/lower/predicates.go', 'return p.l.provePredicateSeen(declaration.Type(), p.active) == nil', 'return false', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/predicate_helper_return'),
 ('callback-proof', 'internal/lower/predicates.go', 'if accepted, err := l.predicateCallbackContract(node); accepted {', 'if accepted, err := l.predicateCallbackContract(node); accepted && false {', './internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/predicate_callback_contract'),
 ('callback-write', 'internal/lower/predicates_contracts.go', 'if n.Kind == ast.KindIdentifier && l.symbol(n) == parameterSymbol && ast.IsAssignmentTarget(n) {', 'if false && n.Kind == ast.KindIdentifier && l.symbol(n) == parameterSymbol && ast.IsAssignmentTarget(n) {','./internal/lower','TestPredicateContractRefusals/reassigned_callback_parameter'),
 ('a-refusal', 'internal/lower/predicates_contracts.go', 'if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {', 'if false && !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {','./internal/oracle','TestPredicateCheckedNarrowing'),
 ('true-check', 'internal/lower/predicates_contracts.go', 'cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', 'if l.checker.TypeToString(target) == "Expression" { return value, nil }; cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', './internal/oracle','TestPredicateCheckedNarrowing/true'),
 ('callback-hatch', 'internal/lower/predicates_contracts.go', 'l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type())', 'false && l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type())', './internal/oracle','TestPredicateCheckedNarrowing/callback'),
 ('false-check', 'internal/lower/predicates_contracts.go', 'cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', 'if l.checker.TypeToString(target) == "Statement" { return value, nil }; cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', './internal/oracle','TestPredicateCheckedNarrowing/false'),
]
for name,path,old,new,pkg,test in mutants:
 p=root/path; original=p.read_text(); assert original.count(old)==1,(name,original.count(old))
 try:
  p.write_text(original.replace(old,new))
  with open('/tmp/predicate-mutant-'+name+'.log','w') as log:
   result=subprocess.run(['go','test',pkg,'-run',test,'-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  print(name,result.returncode,flush=True)
  if result.returncode==0: raise RuntimeError('surviving mutant '+name)
 finally: p.write_text(original)
