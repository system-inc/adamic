"""Built identities and rest-mapping mutants must fail independent outcomes."""
import pathlib,subprocess
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-generic-callable-mutants');scratch.mkdir(exist_ok=True)
source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
needle='''    node := function.Node
    if function.Skipped'''
mutants={
 'drop-generic-declaration-identity':(needle,'''    node := function.Node
    if node.Kind == ast.KindFunctionDeclaration && len(node.TypeParameters()) != 0 { return latentShapeUnknown{Category:"flow the graph can't see",Reason:"mutant dropped generic declaration"} }
    if function.Skipped'''),
 'drop-generic-arrow-identity':(needle,'''    node := function.Node
    if node.Kind == ast.KindArrowFunction && len(node.TypeParameters()) != 0 { return latentShapeUnknown{Category:"flow the graph can't see",Reason:"mutant dropped generic arrow"} }
    if function.Skipped'''),
 'ignore-callback-rest-array':('''            return latentShapeUnknown{Category:"flow the graph can't see", Reason:"callback rest arguments require an array allocation at "+m.Loaded.Where(node)}''','''            continue'''),
}
for name,(old,new) in mutants.items():
 assert source.count(old)==1,(name,source.count(old))
 template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new))
 overlay=scratch/name;binary=scratch/(name+'-binary')
 with (root/'stage3/shape-conformance/overnight'/(name+'.log')).open('w') as log:
  assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0,name+' did not compile'
  status=subprocess.run(['python3','stage3/shape-conformance/generic-callables/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  assert status!=0,name+' escaped'
 print(name+': valid measurement binary; semantic controls caught mutant',flush=True)
