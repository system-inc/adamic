"""Mutation-check measurement guards without changing production or templates."""
import pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path(sys.argv[1]);scratch.mkdir(parents=True,exist_ok=True)
source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'latent-ignore-array-contract':source.replace('if allocation.Kind == "array" {','if false {'),
 'latent-frontier-budget':source.replace('if steps >= 100000 {','if steps >= 0 {'),
 'latent-drop-diagnostic-provenance':source.replace('counts := map[string]map[string]int{}','for i:=range sites { sites[i].DiagnosticCauses=nil }; counts := map[string]map[string]int{}'),
 'latent-ignore-field-type':source.replace('if !checked.IsTypeAssignableTo(declared, expected) {','if false {'),
 'latent-ignore-readiness':source.replace('if staged {\n\t\t\t\t\tready = false','if false {\n\t\t\t\t\tready = false'),
 'latent-ignore-diagnosed-body':source.replace('skipped := len(program.LatentDiagnosticsIn(node.Body())) > 0','skipped := false').replace('ast.IsFunctionLike(parent) && len(program.LatentDiagnosticsIn(parent.Body())) > 0','false'),
}
for name,changed in mutants.items():
 assert changed!=source,name+' did not mutate'
 template=scratch/(name+'.go.txt');template.write_text(changed);overlay=scratch/name;binary=scratch/(name+'-binary');result=scratch/(name+'-result.json')
 with (root/'stage3/shape-conformance/logs'/f'{name}-build.log').open('w') as log:
  assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0
 with (root/'stage3/shape-conformance/logs'/f'{name}.log').open('w') as log:
  assert subprocess.run([str(binary),'/tmp/shape-latent-fixtures','/tmp/shape-latent-fixture-sites.json',str(result)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  status=subprocess.run(['python3','stage3/shape-conformance/latent/audit.py',str(result),'/tmp/shape-latent-fixture-sites.json','/tmp/shape-latent-fixtures','--fixtures'],stdout=log,stderr=subprocess.STDOUT).returncode
  assert status!=0,name+' escaped independent control assertion'
 print(name,'valid measurement binary; independent audit caught mutant',flush=True)
