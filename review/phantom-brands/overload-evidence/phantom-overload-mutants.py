import json,pathlib,subprocess
root=pathlib.Path('/workspace/adamic')
helper='internal/lower/phantom_overload_results.go'
mutants=[
 ('non-brand-result-accepted',helper,'func (l *lowering) phantomOverloadResult(where, implementation *ast.Node, from, to *checker.Type) bool {','func (l *lowering) phantomOverloadResult(where, implementation *ast.Node, from, to *checker.Type) bool {\n if !l.hasPhantom(from) && !l.hasPhantom(to) {return true}', './internal/lower','TestPhantomOverloadLiteralResultRefused'),
 ('ordinary-result-proof-skipped',helper,'l.provenTypesRelation(overload, implementation.Name(), from, to) == nil','true','./internal/lower','TestPhantomOverloadLiteralResultRefused'),
 ('branded-literal-accepted',helper,'func (l *lowering) phantomOverloadResult(where, implementation *ast.Node, from, to *checker.Type) bool {','func (l *lowering) phantomOverloadResult(where, implementation *ast.Node, from, to *checker.Type) bool {\n if l.hasPhantom(to) {return true}', './internal/lower','TestPhantomOverloadBrandLiteralConstraintRefused'),
 ('runtime-brand-result',helper,'return converted, nil','if l.hasPhantom(result) { return ir.Defined{Value:converted, Message:"mutant overload brand"}, nil }; return converted, nil','./internal/lower','TestPhantomOverloadResultCastsAreErased'),
 ('overload-parameter-proof-skipped',helper,'if !l.sameKeeping(fromParameter, toParameter, map[[2]*checker.Type]bool{}) || l.provenTypesRelation(overload, implementation.Name(), fromParameter, toParameter) != nil {','if false && (!l.sameKeeping(fromParameter, toParameter, map[[2]*checker.Type]bool{}) || l.provenTypesRelation(overload, implementation.Name(), fromParameter, toParameter) != nil) {','./internal/lower','TestPhantomOverloadParameterProof'),
 ('overload-value-accepted',helper,'return l.notYet(node, "an overloaded function read as a value")','return nil','./internal/lower','TestPhantomOverloadFunctionValueNotYet'),
 ('overload-header-kept','internal/lower/modules.go','statement.Body() == nil && l.overloadImplementation(statement) != nil','false && statement.Body() == nil && l.overloadImplementation(statement) != nil','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/phantom_overload_results.a'),
 ('any-brand-accepted','internal/lower/phantom_brands.go','if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) {','if false && !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) {','./internal/lower','TestPhantomOverloadAnyBrandRefused'),
]
results=[]
for name,file,needle,replacement,package,test in mutants:
 p=root/file; original=p.read_text(); assert original.count(needle)==1,name
 try:
  p.write_text(original.replace(needle,replacement))
  logfile=pathlib.Path('/tmp/phantom-overload-mutant-'+name+'.log')
  with logfile.open('w') as log:
   run=subprocess.run(['go','test',package,'-run',test,'-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  contents=logfile.read_text()
  results.append({'mutant':name,'test':test,'exit':run.returncode,'assertion_failed':'--- FAIL:' in contents and '[build failed]' not in contents})
 finally:p.write_text(original)
pathlib.Path('/tmp/phantom-overload-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(x['exit']==1 and x['assertion_failed'] for x in results),results
