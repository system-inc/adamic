import pathlib, subprocess
root=pathlib.Path('/tmp/adamic-predicates-views')
mutants=[
 ('helper-proof-combined', [('internal/lower/predicates.go','return p.l.proveFlowPredicateSeen(declaration.Type(), p.active) == nil','return false'),('internal/lower/predicates_proof.go','if values, ok := v.summaries[d]; ok {','if values, ok := v.summaries[d]; false && ok {')], '^TestNativeAgreesWithNode$/internal/oracle/testdata/predicate_helper_return.a$'),
 ('callback-proof-admission', [('internal/lower/predicates.go','if l.predicateParameter(node) != nil {\n\t\treturn nil','if l.predicateParameter(node) != nil {\n\t\treturn (&predicateFlowProof{l: l}).refused(node, "callback admission disabled by mutant")')], '^TestNativeAgreesWithNode$/internal/oracle/testdata/predicate_callback_contract.a$'),
 ('callback-hatch-admission', [('internal/lower/predicates.go','if l.predicateParameter(node) != nil {\n\t\treturn nil','if l.predicateParameter(node) != nil {\n\t\treturn (&predicateFlowProof{l: l}).refused(node, "callback admission disabled by mutant")')], '^TestPredicateCheckedNarrowing$/callback$'),
]
for name,edits,test in mutants:
 originals=[]
 try:
  for path,old,new in edits:
   p=root/path; original=p.read_text(); assert original.count(old)==1,(name,path,original.count(old)); originals.append((p,original)); p.write_text(original.replace(old,new))
  with open('/tmp/predicate-views-stack-mutant-'+name+'.log','w') as log:
   r=subprocess.run(['go','test','./internal/oracle','-run',test,'-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  print(name,r.returncode,flush=True)
  if r.returncode==0:raise RuntimeError('surviving mutant '+name)
 finally:
  for p,original in originals:p.write_text(original)
