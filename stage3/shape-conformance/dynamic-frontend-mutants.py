"""Adapter mutants build and fail the independent source controls."""
import pathlib,subprocess
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-dynamic-frontend-mutants');scratch.mkdir(exist_ok=True)
source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'zero-numeric-key':('ir.NumberConstant{Value:value}','ir.NumberConstant{Value:0}'),
 'zero-string-key':('ir.StringConstant{Index:index}','ir.StringConstant{Index:0}'),
 'drop-key-provenance':('pending = append(pending, access.Key)','/* mutant: drop key provenance */'),
 'ignore-indexed-assignment':('ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left.Kind == ast.KindElementAccessExpression','false && ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left.Kind == ast.KindElementAccessExpression'),
 'ignore-indexed-field-effects':('m.Deinitialized[property.Name] || len(graph.projectionIndex().dynamicStores) != 0','m.Deinitialized[property.Name]'),
 'drop-dynamic-read':('return m.materialize(node, shapeDynamicProjection{Receiver:m.expression(access.Expression), Key:m.expression(key)})','return m.unknown(node, "flow the graph can\'t see", "dynamic element key not modeled", text)'),
}
for name,(before,after) in mutants.items():
 assert before in source,name
 changed=scratch/(name+'.go.txt');changed.write_text(source.replace(before,after))
 # Keeping these locals read makes literal mutants semantic, never compile failures.
 if name=='zero-numeric-key':changed.write_text(changed.read_text().replace('value, err := strconv.ParseFloat(node.Text(), 64)','value, err := strconv.ParseFloat(node.Text(), 64)\n        _ = value'))
 if name=='zero-string-key':changed.write_text(changed.read_text().replace('index := len(m.IR.Strings)','index := len(m.IR.Strings)\n        _ = index'))
 overlay=scratch/name;binary=scratch/(name+'-binary')
 logpath=root/'stage3/shape-conformance/logs'/('dynamic-frontend-'+name+'.log')
 with logpath.open('w') as log:
  assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(changed)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0
  status=subprocess.run(['python3','stage3/shape-conformance/dynamic-keys/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
 text=logpath.read_text();assert status!=0 and 'AssertionError' in text,(name,text)
 print(name+': built measurement-only binary; source control assertion caught mutant',flush=True)
