import os, subprocess
from pathlib import Path
refusals=Path("internal/lower/refusals.go")
lower=Path("internal/lower/lower.go")
original=refusals.read_text(); orchestration=lower.read_text()
block="""	for _, module := range modules {
		if err := lowering.refuse(module); err != nil {
			return nil, err
		}
	}
"""
mutants=[("class-merge",refusals,original.replace("if err := l.refuseClassMerge(node); err != nil {", "if err := l.refuseClassMerge(node); false && err != nil {")),
("function-annotation",refusals,original.replace('if l.isLibraryGlobal(reference.TypeName, "Function") {','if false && l.isLibraryGlobal(reference.TypeName, "Function") {')),
("lowering-first",lower,orchestration.replace(block,"").replace("	if lowering.unlowerable != nil {",block+"	if lowering.unlowerable != nil {")),
("record",refusals,original.replace('if l.checker.GetStringIndexType(l.checker.GetTypeAtLocation(node)) != nil {','if false && l.checker.GetStringIndexType(l.checker.GetTypeAtLocation(node)) != nil {')),
("any",refusals,original.replace('	ast.KindAnyKeyword:        {"any", "name the proven type, or use unknown and narrow it"},\n','')),
("eval",refusals,original.replace('if node.Kind == ast.KindIdentifier && l.isLibraryGlobal(node, "eval") {','if false && node.Kind == ast.KindIdentifier && l.isLibraryGlobal(node, "eval") {')),
("expando",refusals,original.replace('if ast.IsExpandoPropertyDeclaration(declaration) {','if false && ast.IsExpandoPropertyDeclaration(declaration) {')),
("new-function",refusals,original.replace('if node.Kind == ast.KindIdentifier && l.isLibraryGlobal(node, "Function") &&','if false && node.Kind == ast.KindIdentifier && l.isLibraryGlobal(node, "Function") &&'))]
mutants += [
("record-reads",refusals,original.replace('return l.notYet(node, "dynamic record reads', 'return nil // mutant\n\t\t\t// return l.notYet(node, "dynamic record reads')),
("record-writes",refusals,original.replace('return l.notYet(node, "dynamic record writes', 'return nil // mutant\n\t\t\t\t// return l.notYet(node, "dynamic record writes')),
("record-values",refusals,original.replace('(name == "values" || name == "entries")', '(name == "entries")')),
("record-entries",refusals,original.replace('(name == "values" || name == "entries")', '(name == "values")'))]
for name,path,mutant in mutants:
 try:
  if mutant == path.read_text(): raise RuntimeError("unchanged mutant "+name)
  path.write_text(mutant)
  with open("/tmp/refusal-mutant-"+name+".log","w") as log:
   result=subprocess.run(["go","test","./internal/lower","-run","^TestRefusalPassRulings$","-count=1"],stdout=log,stderr=subprocess.STDOUT,env={**os.environ,"ADAMIC_GATE_UNCACHED":"1"})
  print(name,"exit="+str(result.returncode),flush=True)
  if result.returncode == 0: raise RuntimeError("survived mutant "+name)
 finally:
  refusals.write_text(original); lower.write_text(orchestration)
