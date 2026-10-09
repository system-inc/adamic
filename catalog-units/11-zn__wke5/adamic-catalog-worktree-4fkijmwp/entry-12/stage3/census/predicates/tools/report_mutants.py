from pathlib import Path
import subprocess
mutants=[
 ('drop-false-report','cmd/adamic/checks.go','for _, direction := range site.Directions {','for _, direction := range site.Directions {\n if direction.Direction == "false" { continue }','./cmd/adamic','^TestExplainChecksOutput/overload_some_empty$','explanation:'),
 ('trust-inferred-callback','internal/lower/predicates.go','if !l.proveInferredPredicate(implementation, target) {','if false && !l.proveInferredPredicate(implementation, target) {','./internal/lower','^TestPredicateCallbackContracts/inferred_opaque$','want pinned callback argument refusal'),
]
for name,file,before,after,package,test,expected in mutants:
 p=Path('/workspace/adamic')/file;original=p.read_text()
 assert original.count(before)==1,(name,original.count(before))
 try:
  p.write_text(original.replace(before,after))
  log=Path('/tmp/predicate-report-mutant-'+name+'.log')
  with log.open('w') as out:
   result=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; go test '+package+' -run '+test+' -count=1'],cwd='/workspace/adamic',stdout=out,stderr=subprocess.STDOUT)
  output=log.read_text()
  assert result.returncode!=0 and expected in output and '[build failed]' not in output,(name,result.returncode,output)
  print(name+': pinned semantic assertion caught mutant',flush=True)
 finally:p.write_text(original)
