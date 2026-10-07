"""Build valid measurement mutants; independent controls hold generic joins."""
import pathlib
import subprocess

root = pathlib.Path.cwd()
scratch = pathlib.Path('/tmp/shape-generic-mutants')
scratch.mkdir(exist_ok=True)
source = (root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
join = source.replace('Current        int', 'SeenGenericCalls map[int]bool\n\tCurrent        int')
needle = 'func (m *latentShapeMeter) directAllocationCall(node *ast.Node, function latentShapeFunction) ir.Expression {'
join = join.replace(needle, needle + '''
 if len(function.Node.TypeParameters()) != 0 {
  if m.SeenGenericCalls == nil { m.SeenGenericCalls=map[int]bool{} }
  if m.SeenGenericCalls[function.Index] { return ir.NumberConstant{} }
  m.SeenGenericCalls[function.Index]=true
 }
''')
mutants = {
 'generic-drop-later-caller': join,
 'generic-ignore-rest-array': source.replace('if parameter.AsParameterDeclaration().DotDotDotToken != nil {', 'if parameter.AsParameterDeclaration().DotDotDotToken != nil && false {'),
 'generic-ignore-opaque-callee-inputs': source.replace('m.append(ir.Assign{Local:parameter, Value:boundary})', 'm.append(ir.Evaluate{Value:ir.Read{Local:parameter,Of:ir.Object}})'),
 'generic-drop-diagnosed-provenance': source.replace('if function.Skipped {\n\t\t\t\t\t\treturn m.opaqueCallbackCall', 'if function.Skipped && len(declaration.TypeParameters()) == 0 {\n\t\t\t\t\t\treturn m.opaqueCallbackCall'),
}
for name, changed in mutants.items():
 assert changed != source
 template = scratch/(name+'.go.txt'); template.write_text(changed)
 overlay = scratch/name; binary = scratch/(name+'-binary'); result = scratch/(name+'-result.json')
 with (root/'stage3/shape-conformance/logs'/f'{name}.log').open('w') as log:
  assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run([str(binary),'/tmp/shape-latent-fixtures','/tmp/shape-latent-fixture-sites.json',str(result)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  status=subprocess.run(['python3','stage3/shape-conformance/latent/audit.py',str(result),'/tmp/shape-latent-fixture-sites.json','/tmp/shape-latent-fixtures','--fixtures'],stdout=log,stderr=subprocess.STDOUT).returncode
  assert status != 0, name+' escaped independent controls'
 print(name+': valid measurement binary; independent audit caught mutant',flush=True)
