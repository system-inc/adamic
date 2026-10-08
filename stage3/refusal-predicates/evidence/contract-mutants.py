import pathlib, subprocess
root=pathlib.Path('/workspace/adamic')
mutants=[
 ('callback-hatch', 'internal/lower/predicates_contracts.go', 'l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type())', 'false && l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type())', './internal/oracle','TestPredicateCheckedNarrowing/callback'),
 ('open-union', 'internal/lower/predicates_contracts.go', 'if source.Flags()&checker.TypeFlagsUnion == 0 || !l.castUnionWrites(source) {\n\t\treturn castProof{}, false', 'if source.Flags()&checker.TypeFlagsUnion == 0 || !l.castUnionWrites(source) {\n\t\treturn castProof{field: "kind", allowed: []*checker.Type{target}}, true', './internal/oracle', 'TestPredicateOpenContractsStayRefused/open_payload'),
 ('exact-member', 'internal/lower/predicates_contracts.go', 'if !found {\n\t\t\treturn castProof{}, false', 'if !found {\n\t\t\treturn castProof{field: "kind", allowed: []*checker.Type{target}}, true', './internal/oracle', 'TestPredicateOpenContractsStayRefused/extra_refinement'),
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
