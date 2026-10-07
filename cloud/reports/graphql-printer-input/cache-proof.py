import importlib.util,os,pathlib,shutil,tempfile
repo=pathlib.Path('/workspace/adamic');s=importlib.util.spec_from_file_location('npm',repo/'cloud/setup-gate-npm.py');m=importlib.util.module_from_spec(s);s.loader.exec_module(m)
with tempfile.TemporaryDirectory(prefix='graphql-input-',dir='/tmp/adamic-gate') as temporary:
 source=pathlib.Path(temporary)/'source';shutil.copytree(repo/'cloud/gate-inputs/css-printer',source);dest=pathlib.Path('/tmp/adamic-gate/graphql-printer-shared');node='/tmp/adamic-gate/node-pin-tools/bin/node'
 print(m.prepare(source,dest,node));tree=m.tree_digest(dest);stamp=(dest/'.adamic-stamp').read_bytes()
 os.environ['ADAMIC_GATE_UNCACHED']='1';print(m.prepare(source,dest,node));assert m.tree_digest(dest)==tree and (dest/'.adamic-stamp').read_bytes()==stamp;os.environ.pop('ADAMIC_GATE_UNCACHED')
 lock=source/'package-lock.json';lock.write_bytes(lock.read_bytes()+b'\n');answer=m.prepare(source,dest,node);print('lock changed:',answer);assert answer.startswith('installed (npm ci');assert m.tree_digest(dest)!=tree
 print('restore:',m.prepare(repo/'cloud/gate-inputs/css-printer',dest,node));assert m.tree_digest(dest)==tree
print('PASS: warm skip; uncached byte-identical tree/stamp; changed lock forces reinstall')
