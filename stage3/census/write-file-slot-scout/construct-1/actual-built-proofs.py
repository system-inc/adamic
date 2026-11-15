from pathlib import Path
import subprocess,os,json,shutil,concurrent.futures
r=Path('/workspace/scratch/write-file-construct-run');upstream=Path('/workspace/scratch/write-file-slot-typescript');repo=Path('/workspace/scratch/write-file-construct-proof');env=dict(os.environ,TMPDIR=str(r/'tmp'),NODE_PATH=str(repo/'stage3/api/node_modules'));results=[]
for name in ['actual93-before','actual93-after','actual99-before','actual99-after']:
 tree=r/name
 for folder in ['scripts','lib']:shutil.copytree(upstream/folder,tree/folder,dirs_exist_ok=True)
 for file in ['Herebyfile.mjs','.gulp.js','package.json','package-lock.json']:shutil.copy2(upstream/file,tree/file)
 if not (tree/'node_modules').exists():(tree/'node_modules').symlink_to(upstream/'node_modules')
def build(name):
 with (r/(name+'-build.log')).open('w') as log:p=subprocess.run(['npm','run','build:compiler'],cwd=r/name,env=env,stdout=log,stderr=subprocess.STDOUT)
 return {'name':name,'exit':p.returncode}
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:results=list(pool.map(build,['actual93-before','actual93-after','actual99-before','actual99-after']))
(r/'actual-builds.json').write_text(json.dumps(results,indent=2)+'\n');assert all(x['exit']==0 for x in results),results
for n in ['93','99']:
 directory='93-builtin-imports' if n=='93' else '99-wasm-crypto-import';cmd=['node',str(repo/'stage3/adapt'/directory/'proofs/built-witness.cjs'),str(r/f'actual{n}-before'),str(r/f'actual{n}-after'),str(r/('actual-'+n+'-built-fixture')),str(r/('actual-'+n+'-built-witness.json'))]
 with (r/('actual-'+n+'-built-witness.log')).open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 results.append({'name':n+'-built-witness','command':cmd,'exit':p.returncode});(r/'built-proofs.json').write_text(json.dumps(results,indent=2)+'\n');print(n,p.returncode,flush=True)
