from pathlib import Path
import subprocess,os,json,time
r=Path('/workspace/scratch/write-file-construct-run');repo=Path('/workspace/scratch/write-file-construct-proof');a=repo/'stage3/adapt';tree=r/'lane/adapted-tree';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),TMPDIR=str(r/'tmp'),GOTMPDIR=str(r/'tmp'),GOMEMLIMIT='2GiB');results=[]
while not (r/'lane/patch-set.md').exists() and not (r/'lane.exit').exists():time.sleep(5)
assert (r/'lane/patch-set.md').exists(),'fresh apply failed'
commands=[('stock-and-source-js',['node',r/'checks.cjs']),('public-declarations',['node',r/'declarations.cjs']),('real-body-witness',['node',r/'witness.cjs']),('view-moved-witness',['node',a/'99-correctness-write-file-view/witness.cjs',tree]),('99-require-result-contract',['node',a/'99-wasm-crypto-import/proofs/require-result-contract.cjs',tree]),('99-resolve-host-contract',['node',a/'99-wasm-crypto-import/proofs/resolve-host-contract.cjs',tree]),('99-watch-directory-contract',['node',a/'99-wasm-crypto-import/proofs/watch-directory-contract.cjs',tree]),('99-watch-directory-witness',['node',a/'99-wasm-crypto-import/proofs/watch-directory-witness.cjs']),('adamic-types',[r/'adamic','types',*sorted((tree/'src/compiler').rglob('*.ts'))])]
for label,args in commands:
 with (r/(label+'.log')).open('w') as log:p=subprocess.run(list(map(str,args)),env=env,stdout=log,stderr=subprocess.STDOUT)
 results.append({'name':label,'command':list(map(str,args)),'exit':p.returncode});(r/'checks-exits.json').write_text(json.dumps(results,indent=2)+'\n');print(label,p.returncode,flush=True)
