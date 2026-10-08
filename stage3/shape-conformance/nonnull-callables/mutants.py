"""Transparent references must retain wrong-payload and host proof boundaries."""
import pathlib,subprocess,tempfile
root=pathlib.Path.cwd();source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'drop-nonnull-callable-reference':source.replace('transparent = parent.AsNonNullExpression().Expression == node','transparent = false'),
 'nonnull-trusts-wrong-payload':source.replace('if !checked.IsTypeAssignableTo(declared, expected) {','if false {'),
 'nonnull-erases-host-boundaries':source.replace('if m.hostOrigin(node) {','if false && m.hostOrigin(node) {'),
}
with tempfile.TemporaryDirectory(prefix='shape-nonnull-mutants-') as directory:
 scratch=pathlib.Path(directory)
 for name,changed in mutants.items():
  assert changed!=source,name
  template=scratch/(name+'.go.txt');template.write_text(changed);overlay=scratch/name;binary=scratch/(name+'-binary');log_path=root/'stage3/shape-conformance/overnight'/(name+'.log')
  with log_path.open('w') as log:
   subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT,check=True)
   subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT,check=True)
   status=subprocess.run(['python3','stage3/shape-conformance/nonnull-callables/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  output=log_path.read_text();assert status!=0 and 'AssertionError:' in output and "'outcome':" in output,(name,status,output)
  catcher={'drop-nonnull-callable-reference':'aliasIdentity','nonnull-trusts-wrong-payload':'wrongIdentity','nonnull-erases-host-boundaries':'hostIdentity'}[name]
  assert "AssertionError: ('"+catcher+"'" in output,(name,output)
  print(name+': valid built mutant; independent '+catcher+' assertion catches it',flush=True)
