import pathlib, subprocess
root=pathlib.Path('/tmp/adamic-predicates-views')
mutants=[
 ('helper-proof', 'internal/lower/predicates.go', 'return p.l.proveFlowPredicateSeen(declaration.Type(), p.active) == nil', 'return false', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/predicate_helper_return'),
 ('callback-proof', 'internal/lower/predicates.go', 'if accepted, err := l.predicateCallbackContract(node); accepted {', 'if accepted, err := l.predicateCallbackContract(node); accepted && false {', './internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/predicate_callback_contract'),
 ('callback-write', 'internal/lower/predicates_contracts.go', 'if n.Kind == ast.KindIdentifier && l.symbol(n) == parameterSymbol && ast.IsAssignmentTarget(n) {', 'if false && n.Kind == ast.KindIdentifier && l.symbol(n) == parameterSymbol && ast.IsAssignmentTarget(n) {','./internal/lower','TestPredicateContractRefusals/reassigned_callback_parameter'),
 ('a-refusal', 'internal/lower/predicates_contracts.go', 'if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {', 'if false && !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {','./internal/oracle','TestPredicateCheckedNarrowing'),
 ('true-check', 'internal/lower/predicates_contracts.go', 'cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', 'if l.checker.TypeToString(target) == "Expression" { return value, nil }; cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', './internal/oracle','TestPredicateCheckedNarrowing/true'),
 ('callback-hatch', 'internal/lower/predicates_contracts.go', 'l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type())', 'false && l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type())', './internal/oracle','TestPredicateCheckedNarrowing/callback'),
 ('false-check', 'internal/lower/predicates_contracts.go', 'cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', 'if l.checker.TypeToString(target) == "Statement" { return value, nil }; cast := ir.CheckedCast{Value: value, Field: proof.field, Message:', './internal/oracle','TestPredicateCheckedNarrowing/false'),
]
mutants.extend([
 ('view-admission', 'internal/lower/predicates.go', '\t\t_, admission := l.admitPredicateView(node, nil, l.concrete(l.checker.GetTypeFromTypeNode(node.AsTypePredicateNode().Type)))\n\t\treturn admission', '\t\treturn nil', './internal/oracle', '^TestPredicateStructuralViews$/.*invalid$'),
 ('view-a-refusal', 'internal/lower/predicates_views.go', 'if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {', 'if false && !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {', './internal/oracle', '^TestPredicateStructuralViews$'),
 ('view-writable', 'internal/lower/predicates_views.go', '&& l.widened(target, source, map[[2]*checker.Type]bool{}) == nil', '', './internal/oracle', '^TestPredicateWritableViewStaysRefused$'),
])
for name,path,old,new,pkg,test in mutants:
 p=root/path; original=p.read_text(); assert original.count(old)==1,(name,original.count(old))
 secondary=[]
 if name=='view-admission':
  q=root/'internal/lower/predicates_contracts.go'; text=q.read_text(); needle='admitted, err := l.structuralViewCast(node, value, l.checker.GetNonNullableType(l.concrete(source)), l.concrete(target))'; assert text.count(needle)==1
  secondary.append((q,text,text.replace(needle,'return value, nil\n\t\t'+needle)))
 if name=='view-writable':
  q=root/'internal/lower/view_objects.go'; text=q.read_text(); needle='if err := l.widened(target, source, map[[2]*checker.Type]bool{}); err != nil {'; assert text.count(needle)==1
  secondary.append((q,text,text.replace(needle,'if err := l.widened(target, source, map[[2]*checker.Type]bool{}); false && err != nil {')))
 try:
  p.write_text(original.replace(old,new))
  for q,text,changed in secondary:q.write_text(changed)
  with open('/tmp/predicate-views-stack-mutant-'+name+'.log','w') as log:
   result=subprocess.run(['go','test',pkg,'-run',test,'-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  print(name,result.returncode,flush=True)
  if result.returncode==0: print('SURVIVED '+name,flush=True)
 finally:
  p.write_text(original)
  for q,text,changed in secondary:q.write_text(text)
