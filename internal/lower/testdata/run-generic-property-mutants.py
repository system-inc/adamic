import pathlib, subprocess, json
root=pathlib.Path(__file__).resolve().parents[3]
cases=[
('instantiation','generic_closures.go','l.typeMapper = newTypeMapper(sources, arguments)','l.typeMapper = nil\n\tfor _, source := range sources { delete(l.substitution,source) }','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/generic-function-property','Lower:'),
('conversion','expression.go','if (fromGeneric != nil) != (toGeneric != nil) {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('unseeded','generic_closures.go','if _, seeded := families[key]; !seeded {','if _, seeded := families[key]; false && !seeded {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('outer-instantiation','generic_closures.go','if l.typeMapper != nil {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('mixed-union','expression.go','if l.genericSignature(proven) == nil {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('optional-call','generic_closures.go','if node.Flags&ast.NodeFlagsOptionalChain != 0 {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('hoisting','functions.go','l.registerNestedGenerics(body)','if false { l.registerNestedGenerics(body) }','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_properties','Lower:'),
('slot-key','generic_closures.go','return strings.Join(keys, ";")','return strings.Join(keys[:1], ";")','./internal/lower','TestGenericPropertyHasConcreteBodies','want number/string/object bodies'),
('spread','expression.go','if parent != nil && parent.Kind == ast.KindSpreadAssignment && l.genericSignature(l.checker.GetTypeAtLocation(node)) != nil {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('captured-locals','generic_closures.go','l.locals[symbol] = local','if l.result.Locals[local].Global { l.locals[symbol] = local }','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_properties','Lower:'),
('property-observation','expression.go','if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node && l.genericSignature(l.checker.GetTypeAtLocation(node)) != nil {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('erased-cast','expression.go','if assertion.Kind == ast.KindAsExpression && l.genericSignature(l.checker.GetTypeAtLocation(assertion.AsAsExpression().Expression)) != nil && l.genericSignature(l.checker.GetTypeAtLocation(assertion)) == nil {','if false {','./internal/lower','TestGenericPropertyUnsupportedUsesStayDiagnostic','want a diagnostic, got <nil>'),
('typeof','expression.go','if l.genericSignature(l.checker.GetTypeAtLocation(written)) != nil {','if false {','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_properties','stdout differs'),
]
for name,filename,old,new,package,test,expected in cases:
 source=root/'internal/lower'/filename
 original=source.read_text();assert original.count(old)==1,(name,original.count(old))
 mutant=pathlib.Path('/tmp/generic-property-mutant-'+name+'.go');replacement=original.replace(old,new,1)
 if name=='erased-cast': replacement=replacement.replace('assertion := ast.SkipParentheses(node)','assertion := ast.SkipParentheses(node); _ = assertion',1)
 if name=='conversion': replacement=replacement.replace('if fromGeneric != nil && (','if fromGeneric != nil && toGeneric != nil && (',1)
 mutant.write_text(replacement)
 overlay=pathlib.Path('/tmp/generic-property-mutant-'+name+'.json');overlay.write_text(json.dumps({'Replace':{str(source):str(mutant)}}))
 log=pathlib.Path('/tmp/generic-property-mutant-'+name+'.log')
 with log.open('w') as out:
  result=subprocess.run(['go','test','-overlay='+str(overlay),package,'-run',test,'-count=1','-timeout','30m'],cwd=root,stdout=out,stderr=subprocess.STDOUT)
 output=log.read_text();assert result.returncode==1 and expected in output,(name,result.returncode,output)
 assert source.read_text()==original
 print(name,'caught:',expected,flush=True)
