"""Iteration frontier causes and iterable provenance must remain observable."""
import pathlib,subprocess,tempfile
root=pathlib.Path.cwd();source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'erase-for-of-frontier':('reason = "for-of binding requires an iterator value allocation"','reason = "missing expression"'),
 'erase-for-in-frontier':('reason = "for-in binding requires a key producer"','reason = "missing expression"'),
 'drop-iteration-source-dependency':('Dependencies:[]ir.Expression{m.expression(loop.AsForInOrOfStatement().Expression)}','Dependencies:[]ir.Expression{}'),
 'collection-is-iteration-value':('if reason != "" {\n            return latentShapeUnknown','if reason != "" {\n            return m.expression(loop.AsForInOrOfStatement().Expression)\n            return latentShapeUnknown'),
}
with tempfile.TemporaryDirectory(prefix='shape-iteration-mutants-') as directory:
 scratch=pathlib.Path(directory)
 for name,(old,new) in mutants.items():
  assert source.count(old)==1,(name,source.count(old))
  template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new));overlay=scratch/name;binary=scratch/(name+'-binary');log_path=root/'stage3/shape-conformance/overnight'/(name+'.log')
  with log_path.open('w') as log:
   subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT,check=True)
   subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT,check=True)
   status=subprocess.run(['python3','stage3/shape-conformance/iteration-frontiers/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  output=log_path.read_text();assert status!=0 and 'AssertionError:' in output and "'outcome':" in output,(name,status,output)
  print(name+': valid built mutant caught by iteration cause/provenance assertion',flush=True)
