"""A missing result cannot be omitted; terminating positive paths stay precise."""
import pathlib,subprocess,tempfile
root=pathlib.Path.cwd();source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'erase-explicit-return-cause':('Reason:"explicit return without a result at "+m.Loaded.Where(node)','Reason:"missing expression"',True),
 'drop-implicit-result':('if !allocationBodyCannotFallThrough(node.Body()) {','if false && !allocationBodyCannotFallThrough(node.Body()) {',False),
 'one-if-arm-is-enough':('return allocationBodyCannotFallThrough(branch.ThenStatement) && allocationBodyCannotFallThrough(branch.ElseStatement)','return allocationBodyCannotFallThrough(branch.ThenStatement) || allocationBodyCannotFallThrough(branch.ElseStatement)',False),
 'all-blocks-return':('case ast.KindBlock:\n        for _, statement := range node.AsBlock().Statements.Nodes {','case ast.KindBlock:\n        return true\n        for _, statement := range node.AsBlock().Statements.Nodes {',False),
 'ignore-return-sequence':('if allocationBodyCannotFallThrough(statement) { return true }','if false && allocationBodyCannotFallThrough(statement) { return true }',True),
 'throw-is-fallthrough':('case ast.KindReturnStatement, ast.KindThrowStatement:','case ast.KindReturnStatement:',True),
}
with tempfile.TemporaryDirectory(prefix='shape-missing-results-mutants-') as directory:
 scratch=pathlib.Path(directory)
 for name,(old,new,positive) in mutants.items():
  assert source.count(old)==1,(name,source.count(old))
  template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new))
  overlay=scratch/name;binary=scratch/(name+'-binary')
  log_path=root/'stage3/shape-conformance/overnight'/(name+'.log')
  with log_path.open('w') as log:
   assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
   assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0,name+' did not compile'
   status=subprocess.run(['python3','stage3/shape-conformance/missing-results/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
   assert status!=0,name+' escaped'
  output=log_path.read_text()
  assert 'AssertionError:' in output,name+' failed for an unrelated reason'
  outcome='unknown' if positive else 'conforms and ready (free)'
  assert "'outcome': '"+outcome+"'" in output,name+' did not reach its semantic counterexample'
  print(name+': valid measurement binary; independent Node-backed outcome assertion catches mutant',flush=True)
