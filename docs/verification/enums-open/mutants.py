import pathlib, subprocess, json
root=pathlib.Path(__file__).resolve().parents[3]
mutants=[
 ('remove-never-check','internal/lower/enum_never.go','Body: []ir.Statement{ir.Panic{Message: message}}','Body: []ir.Statement{ir.Evaluate{Value: message}, ir.Return{Value: fit(read, of)}}','./internal/oracle','^TestNumericEnumNeverPinned$','exit 0, stdout'),
 ('erase-without-proof','internal/lower/enums.go','return l.enumSwitchCovered(node) && (!l.numericEnum(l.enumIdentity(proven)) || l.enumMemberOrigin(statement.Expression, l.enumIdentity(proven), map[*ast.Node]bool{}))','return l.enumSwitchCovered(node) && proven != nil && statement.Expression != nil','./internal/oracle','^TestNumericEnumNeverPinned$','exit 0, stdout'),
 ('close-numeric-domain','internal/lower/enums.go','if l.numericEnum(l.enumIdentity(to)) {','if false && l.numericEnum(l.enumIdentity(to)) {','./internal/lower','^TestNumericEnumsAreOpen/arithmetic$','unproven value'),
 ('invert-object-presence','internal/lower/control.go','return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: condition}}, nil','return ir.IsUndefined{Value: condition}, nil','./internal/oracle','TestNativeAgreesWithNode/stage3/fixtures/enums/05','exit codes differ'),
 ('permit-enum-object-write','internal/lower/enums.go','if node.Kind == ast.KindCallExpression {','if false && node.Kind == ast.KindCallExpression {','./internal/lower','^TestEnumSlotViews/object_assign$','want Refused'),
 ('repeat-switch-effects','internal/lower/object.go','value = ir.Read{Local: local, Of: value.Type()}','value = ir.Expression(value)','./internal/oracle','^TestNumericEnumNeverPathsPinned/implicit$','stdout differs'),
 ('omit-never-increment-check','internal/lower/assignments.go','if l.enumNeverIdentity(operand, map[*ast.Node]bool{}) != nil {','if false && l.enumNeverIdentity(operand, map[*ast.Node]bool{}) != nil {','./internal/oracle','^TestNumericEnumNeverPathsPinned/update$','exit codes differ'),
 ('permit-member-number','internal/lower/enums.go','if !l.openNumericEnumType(to) {','if false && !l.openNumericEnumType(to) {','./internal/lower','^TestNumericEnumLiteralPromises/member_tag$','want literal-promise refusal'),
 ('trust-exclusion-literal','internal/lower/invariance.go','if l.openNumericEnumType(declared) && !l.enumMemberOrigin','if false && l.openNumericEnumType(declared) && !l.enumMemberOrigin','./internal/lower','^TestNumericEnumLiteralPromises/narrowed_member$','want literal-promise refusal'),
 ('trust-open-object-tag','internal/lower/enums.go','if l.openNumericEnumType(tag) {','if false && l.openNumericEnumType(tag) {','./internal/lower','^TestNumericEnumLiteralPromises/open_tag$','want literal-promise refusal'),
 ('omit-array-enum-storage','internal/lower/enum_never.go','if node.Kind == ast.KindElementAccessExpression {','if false && node.Kind == ast.KindElementAccessExpression {','./internal/oracle','^TestNumericEnumNeverPathsPinned/index$',"can't lower a value of type never"),
]
results=[]
for name,file,old,new,package,test,expected in mutants:
 path=root/file; before=path.read_bytes(); source=before.decode(); assert source.count(old)==1,(name,source.count(old))
 log=pathlib.Path('/tmp/enums-open-mutant-'+name+'.log')
 try:
  path.write_text(source.replace(old,new,1))
  with log.open('w') as output:
   run=subprocess.run(['go','test',package,'-run',test,'-count=1','-v','-timeout','30m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  text=log.read_text(); caught=run.returncode!=0 and expected in text and '[build failed]' not in text and 'clang -std=' not in text
  results.append({'mutant':name,'test':test,'caught':caught,'exit':run.returncode,'log':str(log)})
  print(name, 'CAUGHT' if caught else 'NOT CAUGHT',flush=True)
 finally: path.write_bytes(before)
pathlib.Path('/tmp/enums-open-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(item['caught'] for item in results), results
