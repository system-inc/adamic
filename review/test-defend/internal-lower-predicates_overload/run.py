import pathlib,subprocess,json,time,os
root=pathlib.Path('/workspace/adamic'); tmp=pathlib.Path('/tmp/defend042'); out=root/'review/test-defend/internal-lower-predicates_overload'; out.mkdir(parents=True,exist_ok=True)
plan=[
 dict(id='D1',row='TestIndirectPredicateOverloadIsPending',file='internal/lower/predicates.go',old='"an indirect call of a checked predicate overload"',new='"a direct call of a checked predicate overload"',kind='change constant',aim='exclusive indirect-call capability diagnostic'),
 dict(id='A1',row='TestConditionAssertionAdmission',file='internal/lower/expression.go',old='if prefix.Operator == ast.KindExclamationToken {\n\t\treturn ir.Unary{Operator: ir.Not, Operand: censusCondition(operand)}, nil',new='if prefix.Operator == ast.KindExclamationToken && operand.Type() != ir.Union {\n\t\treturn ir.Unary{Operator: ir.Not, Operand: censusCondition(operand)}, nil',kind='flip condition for boxed operands',aim='runtime negation of an unknown assertion argument; proof-only subsumer does not emit it'),
 dict(id='E1',row='TestEveryNeedsCallbackEffects',file='internal/lower/predicates.go',old='return nil, p.refused(node, fmt.Sprintf("return paths through %s are not verified", node.Kind))',new='return nil, nil',kind='return early',aim='unsupported loop loses all remaining paths, incorrectly certifying the array callback'),
 dict(id='E2',row='TestEveryNeedsCallbackEffects',file='internal/lower/predicates_proof.go',old='return proof, predicateFailure(l, node, "the target has an unsupported runtime contract")',new='return proof, nil',kind='return early',aim='accept unsupported array target after flow proof failed on callback loop'),
 dict(id='E3',row='TestEveryNeedsCallbackEffects',file='internal/lower/predicates.go',old='proof, err := l.provePredicate(node)\n\tif err != nil {\n\t\treturn original\n\t}',new='proof, err := l.provePredicate(node)\n\tif err == nil {\n\t\treturn original\n\t}',kind='flip condition',aim='invert summary-failure handoff so failed array callback proof is admitted'),
 dict(id='A2',row='TestConditionAssertionAdmission',file='internal/lower/predicates_proof.go',old='truthiness: true, cells: []string{"truthy", "falsy"}',new='truthiness: false, cells: []string{"truthy", "falsy"}',kind='change option',aim='condition assertion loses truthiness partition'),
 dict(id='A3',row='TestConditionAssertionAdmission',file='internal/lower/expression.go',old='if prefix.Operator == ast.KindExclamationToken {\n\t\treturn ir.Unary{Operator: ir.Not, Operand: censusCondition(operand)}, nil',new='if prefix.Operator != ast.KindExclamationToken {\n\t\treturn ir.Unary{Operator: ir.Not, Operand: censusCondition(operand)}, nil',kind='flip condition',aim='runtime condition negation fails despite a valid body proof')]
for m in plan:
 original=(root/m['file']).read_text(); assert original.count(m['old'])==1,(m['id'],original.count(m['old'])); m['line']=original[:original.index(m['old'])].count('\n')+1
(out/'plan.json').write_text(json.dumps(plan,indent=2)); metadata=[]; defended=set()
for m in plan:
 if m['row'] in defended:continue
 p=root/m['file']; original=p.read_text()
 try:
  p.write_text(original.replace(m['old'],m['new'],1)); (out/(m['id']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',m['file']],cwd=root))
  for phase,command in [('vet','go vet ./internal/lower/'),('matrix',f'ADAMIC_BUILD_CACHE_DIR=/tmp/defend042/cache/{m["id"]} timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .')]:
   start=time.time(); log=tmp/(m['id']+'-'+phase+'.log')
   with log.open('w') as f:res=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; '+command],cwd=root,stdout=f,stderr=subprocess.STDOUT)
   meta=dict(mutant=m['id'],phase=phase,command=command,exit=res.returncode,wall=time.time()-start,log=log.name);metadata.append(meta);(out/'runs.json').write_text(json.dumps(metadata,indent=2))
   if phase=='vet' and res.returncode:raise RuntimeError('vet failed '+m['id'])
  events=[]
  for line in log.read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  failures=[x['Test'] for x in events if x.get('Action')=='fail' and x.get('Test') and '/'not in x['Test']]
  cooked='test timed out' in log.read_text() or res.returncode==124
  if not cooked and failures==[m['row']]:defended.add(m['row'])
  (out/'progress.json').write_text(json.dumps(dict(defended=sorted(defended),last_mutant=m['id'],failures=failures,cooked=cooked),indent=2))
 finally:p.write_text(original)
