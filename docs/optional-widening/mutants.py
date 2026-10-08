from pathlib import Path
import subprocess
p=Path('internal/lower/optional_widening.go'); original=p.read_text()
mutants={
 'structural-as-class': ('property.Flags&ast.SymbolFlagsOptional != 0 && !isClassInstance(source)', 'property.Flags&ast.SymbolFlagsOptional != 0 && false'),
 'skip-argument': ('func (l *lowering) optionalAtSite(node *ast.Node) *optionalWidening {', 'func (l *lowering) optionalAtSite(node *ast.Node) *optionalWidening {\n if node.Parent != nil && node.Parent.Kind == ast.KindCallExpression { return nil }'),
}
for name,(old,new) in mutants.items():
 assert old in original
 try:
  p.write_text(original.replace(old,new,1))
  with open('/tmp/optional-widening-mutant-'+name+'.log','w') as log:
   result=subprocess.run(['go','test','./internal/lower','-run','TestOptionalWideningRefused/(initializer|argument)$','-count=1'],stdout=log,stderr=subprocess.STDOUT)
  print(name,'exit',result.returncode,flush=True)
 finally: p.write_text(original)
