"""Run serially from the repository root. Restore every edited compiler file."""
from pathlib import Path
import subprocess
root=Path.cwd()
mutants=[
 ("1-helper-parameter-index", "internal/lower/predicates_proof.go", """					claim := predicateOfSignature(v.l.checker, v.l.checker.GetSignatureFromDeclaration(d))
					if claim == nil || claim.ParameterIndex() != 0 {
						return 0, predicateFailure(v.l, n, "helper argument 0 does not occupy its predicate parameter; pass the tested value at the helper's predicate parameter index")
					}
""", "", "p(05|06|13)_"),
 ("2-filter-union-representation", "internal/lower/object.go", 'if name == "filter" && element == ir.Union {', 'if false && name == "filter" && element == ir.Union {', "p(22|24)_"),
 ("3-find-result-representation", "internal/lower/object.go", 'if name == "find" && !(element == ir.MaybeNumber && claim.Type().Flags()&checker.TypeFlagsUndefined != 0) {\n\t\t\tresult, err := l.typeOf(node)', 'if false && name == "find" {\n\t\t\tresult, err := l.typeOf(node)', "p(12|21|23)_"),
 ("4-optional-chain-container", "internal/lower/predicates_proof.go", 'references = append(references, container)', 'references = references', "p14_"),
 ("5a-missing-else", "internal/lower/predicates_proof.go", 'return branch.ElseStatement != nil && v.neverStatement(branch.ThenStatement, visiting) && v.neverStatement(branch.ElseStatement, visiting)', 'return v.neverStatement(branch.ThenStatement, visiting) && (branch.ElseStatement == nil || v.neverStatement(branch.ElseStatement, visiting))', "p26_"),
 ("5b-parameter-write", "internal/lower/predicates_proof.go", 'implementation.Body().ForEachChild(writes)\n\tif changed {', 'implementation.Body().ForEachChild(writes)\n\tif false && changed {', "p27_"),
]
for name,path,old,new,selector in mutants:
 p=root/path;original=p.read_text();assert original.count(old)==1,(name,original.count(old))
 try:
  p.write_text(original.replace(old,new))
  log=Path('/tmp/miscompile-predicates-mutant-'+name+'.log')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestPredicateMiscompileRefusals$/'+selector,'-count=1','-v'],stdout=output,stderr=subprocess.STDOUT,cwd=root)
  print(name,result.returncode,flush=True)
  text=log.read_text()
  if result.returncode==0 or '--- FAIL: TestPredicateMiscompileRefusals/' not in text or '[build failed]' in text:
   raise RuntimeError('mutant not caught by its fixture: '+name)
 finally:p.write_text(original)
