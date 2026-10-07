from pathlib import Path
import subprocess
p=Path('internal/lower/optional_widening.go');original=p.read_text()
mutants={
 'skip-nested': ('if found := l.optionalWidened(l.checker.GetTypeOfSymbol(declared), l.checker.GetTypeOfSymbol(property), nil, visited); found != nil {','if found := (*optionalWidening)(nil); found != nil {','TestOptionalWideningRefused/(nested|generic)$'),
 'reject-class': ('property.Flags&ast.SymbolFlagsOptional != 0 && !isClassInstance(source)','property.Flags&ast.SymbolFlagsOptional != 0','TestOptionalWideningAllowed/class$'),
 'reject-fresh': ('case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:\n\t\t// The literal\'s fields, spreads and elements are checked separately at their own sites.\n\t\treturn nil','case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:\n return l.optionalWidened(l.checker.GetTypeAtLocation(node), target, nil, map[[2]*checker.Type]bool{})','TestOptionalWideningAllowed/fresh$'),
 'skip-spread': ('case ast.KindSpreadAssignment:\n','case ast.KindSpreadAssignment:\n return nil\n','TestOptionalWideningRefused/spread$'),
 'ignore-overwrite-exception': ('skip == nil && !l.checker.IsTypeAssignableTo(source, target)','!l.checker.IsTypeAssignableTo(source, target)','TestOptionalWideningRefused/spread_other_missing$'),
}
for name,(old,new,pattern) in mutants.items():
 assert old in original,name
 try:
  p.write_text(original.replace(old,new,1))
  with open('/tmp/optional-widening-mutant-'+name+'.log','w') as log:
   result=subprocess.run(['go','test','./internal/lower','-run',pattern,'-count=1'],stdout=log,stderr=subprocess.STDOUT)
  print(name,'exit',result.returncode,flush=True)
 finally: p.write_text(original)
