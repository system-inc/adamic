"""Valid constructor dependency mutants lose provenance or invent a result."""
import pathlib,subprocess
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-constructor-mutants');scratch.mkdir(exist_ok=True)
source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
needle='''        return m.materialize(node, latentShapeUnknown{Category:"flow the graph can't see", Reason:"constructor allocation and initialization body not modeled at "+m.Loaded.Where(node), Value:m.text(node), Dependencies:dependencies})'''
mutants={
 'drop-constructor-operand':('dependencies := []ir.Expression{m.expression(constructor.Expression)}','dependencies := []ir.Expression{}'),
 'drop-constructor-arguments':('dependencies = append(dependencies, m.expression(argument))','_ = m.expression(argument)'),
 'constructor-first-argument-is-result':(needle,'if len(dependencies)>1 { return dependencies[1] }\n'+needle),
}
for name,(old,new) in mutants.items():
 assert source.count(old)==1,(name,source.count(old))
 template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new))
 overlay=scratch/name;binary=scratch/(name+'-binary')
 with (root/'stage3/shape-conformance/overnight'/(name+'.log')).open('w') as log:
  assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0,name+' did not compile'
  status=subprocess.run(['python3','stage3/shape-conformance/constructor-frontiers/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  assert status!=0,name+' escaped'
 print(name+': valid measurement binary; semantic controls caught mutant',flush=True)
