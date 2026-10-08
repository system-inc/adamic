"""Valid built mutants must lose the cause or incorrectly erase the missing path."""
import pathlib,subprocess,tempfile
root=pathlib.Path.cwd();source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'erase-uninitialized-declaration-cause':('Reason:"variable declaration without an initializer at "+m.Loaded.Where(node)','Reason:"missing expression"'),
 'drop-uninitialized-declaration':('m.append(ir.Declare{Local: m.local(declaration.Name()), Value: value})','if declaration.Initializer != nil { m.append(ir.Declare{Local: m.local(declaration.Name()), Value: value}) }'),
}
with tempfile.TemporaryDirectory(prefix='shape-uninitialized-mutants-') as directory:
 scratch=pathlib.Path(directory)
 for name,(old,new) in mutants.items():
  assert source.count(old)==1,(name,source.count(old))
  template=scratch/(name+'.go.txt');template.write_text(source.replace(old,new));overlay=scratch/name;binary=scratch/(name+'-binary')
  log_path=root/'stage3/shape-conformance/overnight'/(name+'.log')
  with log_path.open('w') as log:
   subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT,check=True)
   subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT,check=True)
   status=subprocess.run(['python3','stage3/shape-conformance/missing-initializers/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  output=log_path.read_text();assert status!=0 and 'AssertionError:' in output,(name,status)
  assert "'outcome': 'unknown'" in output or "'outcome': 'conforms and ready (free)'" in output,(name,output)
  print(name+': valid built mutant caught by missing-value/cause assertion',flush=True)
