"""Recheck the landing candidate's promise, input, storage and empty-array witnesses."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path.cwd(); out=root/'docs/hidden-boundaries/landing'
source=(root/'internal/lower/overload_results.go').read_text()
guard='body = append(body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: test}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})'
mutants=[
 ('drop-indirect-check','overload_results.go',guard,'if closure == nil { '+guard+' }','./internal/oracle','^TestOverloadValues$/^returned$/^liar$','unchecked wrong result'),
 ('skip-single-signature-check','overload_results.go',guard,'if resolved.Declaration() == overload { '+guard+' }','./internal/oracle','^TestOverloadValues$/^narrow$/^liar$','unchecked wrong result'),
 ('erase-TIn-to-Node','overload_visitor_calls.go','l.localTypes[local] = bindings[i]','l.localTypes[local] = l.concrete(l.checker.GetTypeAtLocation(parameter))','./internal/oracle','^TestOverloadVisitors$/^helper$/^liar$','unproven visitor input admitted'),
 ('admit-unproven-invocation','overload_visitor_domain.go','declaration.Body().ForEachChild(scan)\n\treturn result','declaration.Body().ForEachChild(scan)\n\tresult.unproven = nil\n\treturn result','./internal/oracle','^TestOverloadVisitors$/^overloaded-helper$/^liar$','unproven visitor input admitted'),
 ('nonempty-never-array','object.go','literal := ir.ArrayLiteral{Element: element}','literal := ir.ArrayLiteral{Element: element}\n if l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(node)).Flags() & checker.TypeFlagsNever != 0 { literal.Elements = []ir.Expression{ir.NumberConstant{Value: 1}} }','./internal/oracle','^TestNativeAgreesWithNode$/^internal/oracle/testdata/hidden_boundary_never_array[.]a$','stdout'),
 ('drop-extra-binder-guard','census_small.go','if len(servedTypes) > len(declaredTypes) && !l.censusNullableOverloadResult(produced, promised) {\n\t\t\treturn &Refused','if false {\n\t\t\treturn &Refused','./internal/lower','^TestCensusOverloadBinderGuards$/^result$','got <nil>'),
 ('erase-TNode-object-brand','generic_constraint_storage.go','case ir.Object:\n\t\treturn ir.Union, true','case ir.Object:\n\t\treturn ir.Object, true','./internal/lower','^TestHiddenTNodeConstraintRepresentation$/^object$','constraint representation ='),
]
rows=[]
for name,file,needle,replacement,package,test,marker in mutants:
 path=root/'internal/lower'/file; text=path.read_text();assert text.count(needle)==1,(name,text.count(needle))
 with tempfile.TemporaryDirectory(prefix='hidden-boundaries-mutant-') as temp:
  temp=Path(temp); altered=temp/file; altered.write_text(text.replace(needle,replacement)); overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(altered)}}))
  command=['go','test','-overlay',str(overlay),package,'-run',test,'-count=1','-v']
  with (out/(name+'.log.txt')).open('w') as log: run=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  result=(out/(name+'.log.txt')).read_text();caught=run.returncode!=0 and marker in result and '[build failed]' not in result
  rows.append({'mutant':name,'test':test,'caught':caught,'exit':run.returncode});print(rows[-1],flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
