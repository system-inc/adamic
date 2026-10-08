"""Every binding flow/guard mutant builds and fails an independent control."""
import pathlib,subprocess,tempfile
root=pathlib.Path.cwd();source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'drop-binding-boundary-dependencies':('Reason:reason+" at "+m.Loaded.Where(element), Dependencies:dependencies','Reason:reason+" at "+m.Loaded.Where(element)'),
 'drop-object-binding-edge':('m.allocationBindingDeclaration(declaration.Name(), m.allocationDeclarationProducer(node))','m.allocationDeclarationProducer(node)'),
 'bound-name-is-property-key':('Name:key.Text(), Of:ir.Object','Name:name.Text(), Of:ir.Object'),
 'ignore-binding-default':('case binding.Initializer != nil:','case false && binding.Initializer != nil:'),
 'ignore-binding-rest':('case binding.DotDotDotToken != nil:','case false && binding.DotDotDotToken != nil:'),
 'ignore-binding-array-iteration':('case pattern.Kind == ast.KindArrayBindingPattern:','case false && pattern.Kind == ast.KindArrayBindingPattern:'),
 'ignore-binding-nesting':('case !ast.IsIdentifier(name):','case false && !ast.IsIdentifier(name):'),
 'computed-key-is-bound-name':('reason = "computed binding key not certified"','key = name'),
}
with tempfile.TemporaryDirectory(prefix='shape-binding-mutants-') as directory:
 scratch=pathlib.Path(directory)
 for name,(old,new) in mutants.items():
  assert source.count(old)==1,(name,source.count(old))
  template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new));overlay=scratch/name;binary=scratch/(name+'-binary');log_path=root/'stage3/shape-conformance/overnight'/(name+'.log')
  with log_path.open('w') as log:
   subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT,check=True)
   subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT,check=True)
   status=subprocess.run(['python3','stage3/shape-conformance/object-bindings/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  output=log_path.read_text();assert status!=0 and 'AssertionError:' in output,(name,status)
  assert "'outcome':" in output,(name,output)
  print(name+': valid built mutant caught by binding outcome/cause assertion',flush=True)
