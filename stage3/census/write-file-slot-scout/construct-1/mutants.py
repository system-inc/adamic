from pathlib import Path
import subprocess,os,json,time,shutil,hashlib
r=Path('/workspace/scratch/write-file-construct-run');repo=Path('/workspace/scratch/write-file-construct-proof');a=repo/'stage3/adapt';tree=r/'lane/adapted-tree';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),TMPDIR=str(r/'tmp'),GOTMPDIR=str(r/'tmp'),GOMEMLIMIT='2GiB');result=[]
while not (r/'lane/patch-set.md').exists() and not (r/'lane.exit').exists():time.sleep(5)
assert (r/'lane/patch-set.md').exists()
work=r/'construct-mutant';shutil.copytree(tree/'src',work/'src');
for f in ['package.json','package-lock.json']:shutil.copy2(tree/f,work/f)
def run(label,args):
 with (r/(label+'.log')).open('w') as log:p=subprocess.run(list(map(str,args)),env=env,stdout=log,stderr=subprocess.STDOUT)
 result.append({'name':label,'command':list(map(str,args)),'exit':p.returncode});(r/'mutants.json').write_text(json.dumps(result,indent=2)+'\n');print(label,p.returncode,flush=True);return p.returncode
assert run('construct-restoring-mutant',['node',a/'99-correctness-write-file-construct/mutant.cjs',work])==0
actual={f.relative_to(work).as_posix():hashlib.sha256(f.read_bytes()).hexdigest() for f in (work/'src').rglob('*') if f.is_file()};expected={f.relative_to(r/'control-adapted').as_posix():hashlib.sha256(f.read_bytes()).hexdigest() for f in (r/'control-adapted/src').rglob('*') if f.is_file()};assert actual==expected,'mutant does not exactly restore control';(r/'construct-mutant-source-identity.json').write_text(json.dumps({'allSourceFiles':len(actual),'exactParentPlusMainBytes':True,'files':actual},indent=2)+'\n')
assert run('construct-restoring-mutant-replay',['node',a/'99-correctness-write-file-construct/mutant.cjs',work])==0
run('construct-mutant-production-entry',[r/'adamic','js',work/'src/tsc/tsc.ts','--module-format','commonjs-bundle'])
run('99-ternary-mutant-valid-fixture',['node',a/'99-wasm-crypto-import/proofs/ternary-mutant.cjs',r/'prefix99-after'])
for mutant in ['require-result','resolve-host','watch-directory']:
 for mode in (['restore','unnormalized','optional-contract'] if mutant=='watch-directory' else ['restore']):
  shutil.copy2(tree/'src/compiler/sys.ts',work/'src/compiler/sys.ts')
  run('99-'+mutant+'-'+mode+'-mutant',['node',a/'99-wasm-crypto-import'/(mutant+'-mutant.cjs'),work,mode])
  run('99-'+mutant+'-'+mode+'-refused',['node',a/'99-wasm-crypto-import/adapt.cjs',work])
