from pathlib import Path
import subprocess,tarfile,io,json,os,hashlib
r=Path('/workspace/scratch/write-file-construct-run');tree=r/'lane/adapted-tree';repo=Path('/workspace/scratch/write-file-construct-final');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),TMPDIR=str(r/'tmp'));results=[]
def hashes(folder):return {f.relative_to(folder).as_posix():hashlib.sha256(f.read_bytes()).hexdigest() for f in (folder/'src').rglob('*') if f.is_file()}
for n,directory in [('93','93-builtin-imports'),('99','99-wasm-crypto-import')]:
 for state in ['before','after']:
  dest=r/f'actual{n}-{state}';dest.mkdir(exist_ok=True)
  data=subprocess.check_output(['git','-C',str(tree),'archive',f'refs/stage3/{directory}-{state}','src','package.json','package-lock.json'])
  with tarfile.open(fileobj=io.BytesIO(data)) as tar:tar.extractall(dest,filter='data')
  (dest/'node_modules').symlink_to('/workspace/scratch/write-file-slot-typescript/node_modules')
 identity={state:hashes(r/f'actual{n}-{state}')==hashes(r/f'prefix{n}-{state}') for state in ['before','after']}
 results.append({'name':n+'-prefix-context-identity','wholeSourceIdentity':identity})
 commands=[]
 if n=='93' and not all(identity.values()):commands+=[('93-actual-owner-contract',[repo/'stage3/adapt'/directory/'proofs/owner-contract.cjs',r/'actual93-before',r/'actual93-after',r/'93-actual-owner-contract'])]
 commands += [(n+'-actual-bindings',[repo/'stage3/adapt'/directory/'proofs/bindings.cjs',r/f'actual{n}-before',r/f'actual{n}-after',r/(n+'-actual-bindings.json')])]
 if n=='99':commands += [('99-actual-source-transpile-check',[repo/'stage3/adapt'/directory/'proofs/check.cjs',r/'actual99-before',r/'actual99-after',r/'99-actual-checks.json'])]
 for label,args in commands:
  with (r/(label+'.log')).open('w') as log:p=subprocess.run(['node',*map(str,args)],env=env,stdout=log,stderr=subprocess.STDOUT)
  results.append({'name':label,'command':['node',*map(str,args)],'exit':p.returncode});(r/'actual-stage-proofs.json').write_text(json.dumps(results,indent=2)+'\n');print(label,p.returncode,flush=True)
(r/'actual-stage-proofs.json').write_text(json.dumps(results,indent=2)+'\n')
