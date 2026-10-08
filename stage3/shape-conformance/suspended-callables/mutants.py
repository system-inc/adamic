"""Promise and iterator allocations cannot inherit a body's ordinary return proof."""
import pathlib,subprocess,tempfile
root=pathlib.Path.cwd();source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'async-name-is-a-protocol':('if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {','if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) || (node.Kind == ast.KindFunctionDeclaration && node.Name() != nil && node.Name().Text() == "async") {'),
 'ignore-promise-wrapper':('if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {','if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) && false {'),
 'ignore-iterator-wrapper':('if generator {','if generator && false {'),
 'ignore-direct-suspension-protocol':('''    if reason := allocationCallableProtocol(function.Node); reason != "" {
        return m.opaqueAllocationCall(node, function, reason)
    }''',''),
 'ignore-callback-suspension-protocol':('''    if reason := allocationCallableProtocol(node); reason != "" {
        return latentShapeUnknown{Category:"flow the graph can't see", Reason:reason+" at "+m.Loaded.Where(node)}
    }''',''),
}
with tempfile.TemporaryDirectory(prefix='shape-suspended-mutants-') as directory:
 scratch=pathlib.Path(directory)
 for name,(old,new) in mutants.items():
  assert source.count(old)==1,(name,source.count(old))
  template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new))
  overlay=scratch/name;binary=scratch/(name+'-binary')
  log_path=root/'stage3/shape-conformance/overnight'/(name+'.log')
  with log_path.open('w') as log:
   assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
   assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0,name+' did not compile'
   status=subprocess.run(['python3','stage3/shape-conformance/suspended-callables/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
   assert status!=0,name+' escaped'
  output=log_path.read_text()
  assert "AssertionError:" in output,name+' failed for an unrelated reason'
  if name=='async-name-is-a-protocol':
   assert 'synchronousAsyncNameRead' in output and "'outcome': 'unknown'" in output,name+' did not lose a valid synchronous proof'
  else:
   assert "'outcome': 'conforms and ready (free)'" in output,name+' did not falsely certify a wrapper'
  print(name+': valid measurement binary; independent Node-backed outcome assertion catches mutant',flush=True)
