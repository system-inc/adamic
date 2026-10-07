import importlib.util,json,os,pathlib,shutil,subprocess,tempfile
repo=pathlib.Path('/workspace/adamic'); spec=importlib.util.spec_from_file_location('npm',repo/'cloud/setup-gate-npm.py');npm=importlib.util.module_from_spec(spec);spec.loader.exec_module(npm)
root=pathlib.Path(tempfile.mkdtemp(prefix='estree-proof-',dir='/tmp/adamic-gate'));source=root/'source';shutil.copytree(repo/'cloud/gate-inputs/css-printer',source);dest=pathlib.Path('/tmp/adamic-gate/estree-shared');node='/workspace/adamic-tools/bin/node'
print('warm:',npm.prepare(source,dest,node));original=npm.tree_digest(dest)
os.environ['ADAMIC_GATE_UNCACHED']='1';print('uncached:',npm.prepare(source,dest,node));assert npm.tree_digest(dest)==original;os.environ.pop('ADAMIC_GATE_UNCACHED')
lock=source/'package-lock.json';lock.write_bytes(lock.read_bytes()+b'\n');print('changed lock:',npm.prepare(source,dest,node));assert npm.tree_digest(dest)!=original
print('restore:',npm.prepare(repo/'cloud/gate-inputs/css-printer',dest,node));assert npm.tree_digest(dest)==original
inputs=dict(lock='lock',manifest='manifest',bootstrap='bootstrap',helper='helper',node='node')
for dropped in inputs:
 def mutant(**values):
  values.pop(dropped);return npm.digest(json.dumps(values,sort_keys=True).encode())
 try:assert mutant(**inputs)!=mutant(**dict(inputs,**{dropped:'changed'})),dropped
 except AssertionError:print('caught drop-'+dropped+' mutant: key failed to invalidate')
 else:raise AssertionError('mutant survived')
print('PASS: warm, uncached bytes, lock rerun, five key mutants')
